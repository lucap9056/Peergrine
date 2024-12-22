package main

import (
	"log"
	API "peergrine/msg-bridge/api"
	AppConfig "peergrine/msg-bridge/app-config"
	Storage "peergrine/msg-bridge/storage"
	Pulsar "peergrine/utils/pulsar"
	Shutdown "peergrine/utils/shutdown"
	"time"
)

func main() {

	shutdown := Shutdown.New()
	defer shutdown.Shutdown("")

	config, err := AppConfig.Init()
	if err != nil {
		log.Panicln(err)
	}

	{
		var pulsar *Pulsar.Client

		if config.PulsarAddrs != "" {
			pulsar, err = Pulsar.New(config.PulsarAddrs, config.PulsarTopic, config.Id)
			if err != nil {
				log.Println(err)
				return
			}
			defer pulsar.Close()
		}

		storage, err := Storage.New(config.Id, config.RedisAddr)
		if err != nil {
			log.Println(err)
			return
		}

		app, err := API.New(config, storage, pulsar)
		if err != nil {
			log.Println(err)
			return
		}
		defer app.Close()

		go func() {
			err := app.Run()
			if err != nil {
				shutdown.Shutdown("")
			}
		}()
	}

	shutdown.Wait()

	log.Println("Received stop signal or timeout, shutting down...")

	time.Sleep(2 * time.Second)

	log.Println("Service stopped.")
}
