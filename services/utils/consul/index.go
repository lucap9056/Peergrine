package consulclient

import (
	"log"
	"time"

	"github.com/hashicorp/consul/api"
)

const _MAX_RETRY_ATTEMPTS = 10
const _RETRY_INTERVAL_TIME = time.Second * 5

func New(addr string) (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = addr

	var client *api.Client
	var err error

	for attempt := 1; attempt <= _MAX_RETRY_ATTEMPTS; attempt++ {
		client, err = api.NewClient(config)
		if err == nil {
			return client, nil
		}

		log.Printf("Attempt %d: Failed to create Consul client: %v", attempt, err)
		time.Sleep(_RETRY_INTERVAL_TIME)
	}

	return nil, err
}
