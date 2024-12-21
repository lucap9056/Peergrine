package main

import (
	"fmt"
	"log"
	API "peergrine/rtc-bridge/api"
	AppConfig "peergrine/rtc-bridge/app-config"
	Storage "peergrine/rtc-bridge/storage"
	Consul "peergrine/utils/consul"
	ConsulService "peergrine/utils/consul/service"
	Kafka "peergrine/utils/kafka"
	Kafker "peergrine/utils/kafker-client"
	Shutdown "peergrine/utils/shutdown"
	"time"
)

func main() {

	log.Println("Starting the rtc-bridge service...")

	shutdown := Shutdown.New()
	defer shutdown.Shutdown("")

	log.Println("Initializing configuration...")
	config, err := AppConfig.Init()
	if err != nil {
		log.Panicln("Failed to initialize configuration:", err)
	}

	{
		kafkaChannelId := int32(-1)

		if config.KafkerAddr != "" {
			kafker, err := Kafker.New(config.KafkerAddr)
			if err != nil {
				log.Println("Failed to initialize Kafker client:", err)
				return
			}

			serviceId := config.ConsulConfig.ServiceId
			serviceName := config.ConsulConfig.ServiceName
			topicName := config.KafkaTopic

			log.Println("Applying for Kafka partition...")
			partitionId, err := kafker.RequestPartition(serviceId, serviceName, topicName)
			if err != nil {
				log.Println("Failed to apply Kafka partition:", err)
				return
			}
			defer kafker.ReleasePartition(serviceId)
			log.Printf("Kafka partition acquired: %d\n", partitionId)

			kafkaChannelId = partitionId
		}

		var kafka *Kafka.Client

		if config.KafkaAddr != "" {
			log.Println("Kafka address provided, setting up Kafka...")

			log.Println("Initializing Kafka client...")
			kafka, err = Kafka.New(config.KafkaAddr)
			if err != nil {
				log.Println("Failed to initialize Kafka client:", err)
				return
			}

			log.Println("Kafka setup complete.")
		} else {
			log.Println("No Kafka address provided, skipping Kafka setup.")
		}

		log.Println("Setting up storage...")
		storage, err := Storage.New(kafkaChannelId, config.RedisAddr)
		if err != nil {
			log.Println("Failed to initialize storage:", err)
			return
		}
		log.Println("Storage setup complete.")

		log.Println("Initializing API application...")
		app, err := API.New(config, storage, kafka, kafkaChannelId)
		if err != nil {
			log.Println("Failed to initialize API:", err)
			return
		}
		defer app.Close()
		log.Println("API application initialized successfully.")

		go func() {
			log.Println("Running API application...")
			err := app.Run()
			if err != nil {
				shutdown.Shutdown("Failed to connect Auth_gRPC on address %s: %v", config.AuthAddr, err)
			}
		}()
	}

	{
		if config.ConsulAddr != "" {
			log.Println("Consul address provided, setting up Consul service...")

			consulClient, err := Consul.New(config.ConsulAddr)
			if err != nil {
				log.Println("Failed to initialize Consul client:", err)
				return
			}

			if config.ConsulConfig.ServiceAddress == "" {

				log.Println("Fetching local IPv4 address for service registration...")
				serviceAddress, err := ConsulService.GetLocalIPV4Address()
				if err != nil {
					log.Println("Failed to get local IP address:", err)
					return
				}

				config.ConsulConfig.ServiceAddress = serviceAddress

			}

			log.Println("Registering service to Consul...")
			consulService, err := ConsulService.New(consulClient, config.ConsulConfig)
			if err != nil {
				log.Println(err)
				return
			}

			if err := consulService.Register(); err != nil {
				log.Println("Failed to register service to Consul:", err)
				return
			}
			defer consulService.Close()
			log.Println("Service registered to Consul successfully.")

			go func() {
				addr := fmt.Sprintf(":%v", config.ConsulConfig.ServicePort)
				log.Printf("Running Consul service on %s...\n", addr)

				if err := consulService.RunTCP(addr); err != nil {
					shutdown.Shutdown("%v", err)
				}
			}()
		} else {
			log.Println("No Consul address provided, skipping Consul setup.")
		}
	}

	log.Println("Waiting for shutdown signal or timeout...")
	shutdown.Wait()

	log.Println("Received stop signal or timeout, shutting down service...")

	time.Sleep(2 * time.Second)

	log.Println("Service stopped.")
}
