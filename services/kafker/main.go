package main

import (
	"fmt"
	"log"
	"net"
	API "peergrine/kafker/api"
	AppConfig "peergrine/kafker/app-config"
	Cluster "peergrine/kafker/cluster"
	Kafka "peergrine/kafker/kafka"
	Storage "peergrine/kafker/storage"
	Consul "peergrine/utils/consul"
	ConsulManager "peergrine/utils/consul/manager"
	ConsulService "peergrine/utils/consul/service"
	Shutdown "peergrine/utils/shutdown"
	"time"
)

func main() {

	shutdown := Shutdown.New()
	defer shutdown.Shutdown("")

	config, zookeeper, err := AppConfig.Init()
	if err != nil {
		log.Println("Error initializing application configuration:", err)
		return
	}
	defer zookeeper.Close()

	consulClient, err := Consul.New(config.ConsulAddr)
	if err != nil {
		log.Println("Error creating Consul client:", err)
		return
	}

	kafka, err := Kafka.New(config.KafkaAddr)
	if err != nil {
		log.Println("Error connecting to Kafka:", err)
		return
	}
	defer kafka.Close()

	consulManager := ConsulManager.New(consulClient)

	storage, err := Storage.New(zookeeper, kafka, consulManager, config.ClusterMode)
	if err != nil {
		log.Println("Error initializing storage:", err)
		return
	}

	if config.ClusterMode {
		cluster, err := Cluster.New(zookeeper, storage)
		if err != nil {
			log.Println("Error initializing cluster manager:", err)
			return
		}
		defer cluster.Stop()
	}

	app, err := API.New(storage)
	if err != nil {
		log.Println("Error initializing API service:", err)
		return
	}
	defer app.Stop()

	go func() {
		if err := app.Run(config.Addr); err != nil {
			shutdown.Shutdown("API service error: %v", err)
		}
	}()

	if config.ConsulConfig.ServiceAddress == "" {

		serviceAddress, err := getLocalIPV4Address()
		if err != nil {
			log.Println("Error getting local IPv4 address:", err)
			return
		}
		config.ConsulConfig.ServiceAddress = serviceAddress

	}

	service, err := ConsulService.New(consulClient, config.ConsulConfig)
	if err != nil {
		log.Println(err)
		return
	}

	if err := service.Register(); err != nil {
		log.Println("Error registering service with Consul:", err)
		return
	}
	defer service.Deregister()

	go func() {
		addr := fmt.Sprintf(":%v", config.ConsulConfig.ServicePort)

		if err := service.RunTCP(addr); err != nil {
			shutdown.Shutdown("Consul service error: %v", err)
		}
	}()

	shutdown.Wait()

	log.Println("Received stop signal or timeout, shutting down...")

	time.Sleep(2 * time.Second)

	log.Println("Service stopped.")
}

// getLocalIPV4Address retrieves the first reachable IPv4 address of the machine.
func getLocalIPV4Address() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("error getting network interfaces: %w", err)
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return "", fmt.Errorf("error getting addresses for interface %s: %w", iface.Name, err)
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no reachable IPv4 address found")
}
