package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net"
	ClientEndpoint "peergrine/jwtissuer/api/client-endpoint"
	ConnMap "peergrine/jwtissuer/api/conn-map"
	ServiceEndpoint "peergrine/jwtissuer/api/service-endpoint"
	AppConfig "peergrine/jwtissuer/app-config"
	Storage "peergrine/jwtissuer/storage"
	Consul "peergrine/utils/consul"
	ConsulService "peergrine/utils/consul/service"
	Kafka "peergrine/utils/kafka"
	Kafker "peergrine/utils/kafker-client"
	Shutdown "peergrine/utils/shutdown"
	"time"
)

func main() {

	log.Println("Starting application...")
	shutdown := Shutdown.New()
	defer func() {
		log.Println("Initiating shutdown...")
		shutdown.Shutdown("")
	}()

	config, err := AppConfig.Init()
	if err != nil {
		log.Printf("Failed to initialize app config: %v", err)
		return
	}
	log.Println("App configuration initialized successfully.")

	secret, err := generateSecret(32)
	if err != nil {
		log.Printf("Failed to generate secret: %v", err)
		return
	}
	log.Println("Secret generated successfully.")

	storage, err := Storage.New(config.Id, config.RedisAddr)
	if err != nil {
		log.Printf("Failed to initialize storage: %v", err)
		return
	}
	log.Println("Storage initialized successfully.")

	if err := storage.SaveSecret(secret); err != nil {
		log.Printf("Failed to save secret: %v", err)
		return
	}
	log.Println("Secret saved in storage.")

	var kafka *Kafka.Client
	var kafkaChannelId int32

	if config.KafkaAddr != "" {
		kafker, err := Kafker.New(config.KafkerAddr)
		if err != nil {
			log.Println(err)
			return
		}

		serviceId := config.ConsulConfig.ServiceId
		serviceName := config.ConsulConfig.ServiceName
		topicName := config.KafkaTopic

		partitionId, err := kafker.RequestPartition(serviceId, serviceName, topicName)
		if err != nil {
			log.Println(err)
			return
		}
		defer kafker.ReleasePartition(serviceId)

		log.Println("listen kafka partition: ", partitionId)
		kafkaChannelId = partitionId

		kafka, err = Kafka.New(config.KafkaAddr)
		if err != nil {
			log.Println(err)
			return
		}

	}

	connMap := ConnMap.New()

	{
		// Start the service endpoint
		log.Println("Initializing service endpoint...")
		serviceEndpoint := ServiceEndpoint.New(storage, config, connMap, kafka, kafkaChannelId)
		defer serviceEndpoint.Close()
		log.Println("Service endpoint initialized.")

		go func() {
			log.Printf("Running service endpoint on %s...", config.ServiceEndpointAddr)
			if err := serviceEndpoint.Run(config.ServiceEndpointAddr); err != nil {
				log.Printf("Service endpoint failed: %v", err)
				shutdown.Shutdown("Service endpoint failed: %v", err)
			}
		}()
	}

	{
		// Initialize and start the client endpoint
		log.Println("Initializing client endpoint...")
		clientEndpoint, err := ClientEndpoint.New(storage, config, connMap, kafkaChannelId)
		if err != nil {
			log.Printf("Failed to initialize client endpoint: %v", err)
			return
		}
		log.Println("Client endpoint initialized.")

		go func() {
			log.Printf("Running client endpoint on %s...", config.ClientEndpointAddr)
			if err := clientEndpoint.Run(config.ClientEndpointAddr); err != nil {
				log.Printf("Client endpoint failed: %v", err)
				shutdown.Shutdown("Client endpoint failed: %v", err)
			}
		}()
	}

	{
		if config.ConsulAddr != "" {
			log.Printf("Initializing Consul client with address %s...", config.ConsulAddr)
			consulClient, err := Consul.New(config.ConsulAddr)
			if err != nil {
				log.Printf("Failed to initialize Consul client: %v", err)
				return
			}
			log.Println("Consul client initialized.")

			if config.ConsulConfig.ServiceAddress == "" {
				log.Println("Service address not specified in config; fetching local IPv4 address...")
				serviceAddress, err := getLocalIPV4Address()
				if err != nil {
					log.Printf("Failed to get local IPv4 address: %v", err)
					return
				}
				config.ConsulConfig.ServiceAddress = serviceAddress
				log.Printf("Service address set to local IP: %s", serviceAddress)
			}

			log.Println("Initializing Consul service registration...")
			consulService, err := ConsulService.New(consulClient, config.ConsulConfig)
			if err != nil {
				log.Printf("Failed to initialize Consul service: %v", err)
				return
			}
			log.Println("Consul service initialized.")

			if err := consulService.Register(); err != nil {
				log.Printf("Failed to register service with Consul: %v", err)
				return
			}
			log.Println("Service registered with Consul.")
			defer consulService.Close()

			go func() {
				addr := fmt.Sprintf(":%v", config.ConsulConfig.ServicePort)
				log.Printf("Running Consul TCP service on %s...", addr)
				if err := consulService.RunTCP(addr); err != nil {
					log.Printf("Consul service TCP run failed: %v", err)
					shutdown.Shutdown("%v", err)
				}
			}()
		}
	}

	log.Println("Waiting for shutdown signal...")
	shutdown.Wait()

	log.Println("Received stop signal or timeout, shutting down...")
	time.Sleep(2 * time.Second)
	log.Println("Service stopped.")
}

func generateSecret(length int) ([]byte, error) {
	log.Printf("Generating secret of length %d...", length)
	secret := make([]byte, length)
	_, err := rand.Read(secret)
	if err != nil {
		log.Printf("Failed to generate random secret: %v", err)
		return nil, err
	}
	log.Println("Secret generation successful.")
	return secret, nil
}

func getLocalIPV4Address() (string, error) {
	log.Println("Fetching local IPv4 address...")
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Failed to fetch network interfaces: %v", err)
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("Failed to fetch addresses for interface %s: %v", iface.Name, err)
			return "", err
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				log.Printf("Found local IPv4 address: %s", ipnet.IP.String())
				return ipnet.IP.String(), nil
			}
		}
	}

	err = fmt.Errorf("no reachable IPv4 address found")
	log.Println(err)
	return "", err
}
