package main

import (
	"fmt"
	"log"
	API "peergrine/msg-bridge/api"
	AppConfig "peergrine/msg-bridge/app-config"
	Storage "peergrine/msg-bridge/storage"
	Consul "peergrine/utils/consul"
	ConsulService "peergrine/utils/consul/service"
	Kafka "peergrine/utils/kafka"
	Kafker "peergrine/utils/kafker-client"
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

		storage, err := Storage.New(kafkaChannelId, config.RedisAddr)
		if err != nil {
			log.Println(err)
			return
		}

		app, err := API.New(config, storage, kafka, kafkaChannelId)
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

	{
		if config.ConsulAddr != "" {

			consulClient, err := Consul.New(config.ConsulAddr)
			if err != nil {
				log.Println(err)
				return
			}

			if config.ConsulConfig.ServiceAddress == "" {

				serviceAddress, err := ConsulService.GetLocalIPV4Address()
				if err != nil {
					log.Println(err)
					return
				}

				config.ConsulConfig.ServiceAddress = serviceAddress

			}

			consulService, err := ConsulService.New(consulClient, config.ConsulConfig)
			if err != nil {
				log.Println(err)
				return
			}

			if err := consulService.Register(); err != nil {
				log.Println(err)
				return
			}
			defer consulService.Close()

			go func() {
				addr := fmt.Sprintf(":%v", config.ConsulConfig.ServicePort)

				if err := consulService.RunTCP(addr); err != nil {
					shutdown.Shutdown("%v", err)
				}
			}()
		}
	}

	shutdown.Wait()

	log.Println("Received stop signal or timeout, shutting down...")

	time.Sleep(2 * time.Second)

	log.Println("Service stopped.")
}
