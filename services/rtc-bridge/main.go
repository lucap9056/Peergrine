package main

import (
	"log"
	API "peergrine/rtc-bridge/api"
	AppConfig "peergrine/rtc-bridge/app-config"
	Storage "peergrine/rtc-bridge/storage"
	Pulsar "peergrine/utils/pulsar"
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

		var pulsar *Pulsar.Client

		if config.PulsarAddrs != "" {
			pulsar, err = Pulsar.New(config.PulsarAddrs, config.PulsarTopic, config.Id)
			if err != nil {
				log.Println(err)
				return
			}
		}

		log.Println("Setting up storage...")
		storage, err := Storage.New(config.Id, config.RedisAddr)
		if err != nil {
			log.Println("Failed to initialize storage:", err)
			return
		}
		log.Println("Storage setup complete.")

		log.Println("Initializing API application...")
		app, err := API.New(config, storage, pulsar)
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

	log.Println("Waiting for shutdown signal or timeout...")
	shutdown.Wait()

	log.Println("Received stop signal or timeout, shutting down service...")

	time.Sleep(2 * time.Second)

	log.Println("Service stopped.")
}
