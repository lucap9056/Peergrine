package consulmanager

import (
	"github.com/hashicorp/consul/api"
)

type Manager struct {
	*api.Client
}

func New(client *api.Client) *Manager {
	return &Manager{
		client,
	}
}
