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
	Pulsar "peergrine/utils/pulsar"
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

	var pulsar *Pulsar.Client

	if config.PulsarAddrs != "" {
		pulsar, err = Pulsar.New(config.PulsarAddrs, config.PulsarTopic, config.Id)
		if err != nil {
			log.Println(err)
			return
		}
		defer pulsar.Close()
	}

	connMap := ConnMap.New()

	{
		// Start the service endpoint
		log.Println("Initializing service endpoint...")
		serviceEndpoint := ServiceEndpoint.New(storage, config, connMap, pulsar)
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
		clientEndpoint, err := ClientEndpoint.New(storage, config, connMap)
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
