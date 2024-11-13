package consulmanager

import "github.com/hashicorp/consul/api"

func (m *Manager) GetServices(serviceName string) ([]*api.CatalogService, error) {
	services, _, err := m.Catalog().Service(serviceName, "", nil)

	return services, err
}
