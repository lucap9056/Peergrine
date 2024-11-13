package consulclient

import (
	"github.com/hashicorp/consul/api"
)

func New(addr string) (*api.Client, error) {

	config := api.DefaultConfig()
	config.Address = addr

	return api.NewClient(config)
}
