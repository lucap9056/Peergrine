package appconfig

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	Configurator "peergrine/utils/configurator"
	Consul "peergrine/utils/consul/service"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

const (
	_DEFAULT_CLIENT_ENDPOINT_ADDRESS  = ":80"
	_DEFAULT_SERVICE_ENDPOINT_ADDRESS = ":50051"
	_DEFAULT_REDIS_ADDRESS            = "" //redis:6379
	_DEFAULT_BEARER_TOKEN_DURATION    = "3600"
	_DEFAULT_REFRESH_TOKEN_DURATION   = "7200"
	_DEFAULT_CONSUL_ADDRESS           = "" //consul:8500
	_DEFAULT_SERVICE_NAME             = "JWTIssuer"
	_DEFAULT_SERVICE_HEALTHY_PORT     = "4000"
	_DEFAULT_ZK_CONFIG_PATH           = "/jwtissuer"
)

type AppConfig struct {
	Id                   string         `json:"-" config:"APP_ID"`
	ClientEndpointAddr   string         `json:"client_endpoint_address" config:"APP_CLIENTENDPOINT_ADDR"`
	ServiceEndpointAddr  string         `json:"service_endpoint_address" config:"APP_SERVICEENDPOINT_ADDR"`
	RedisAddr            string         `json:"redis_address" config:"APP_REDIS_ADDR"`
	BearerTokenDuration  string         `json:"bearer_token_duration" config:"APP_BEARER_TOKEN_DURATION"`
	RefreshTokenDuration string         `json:"refresh_token_duration" config:"APP_REFRESH_TOKEN_DURATION"`
	ConsulAddr           string         `json:"consul_address" config:"APP_CONSUL_ADDR"`
	ConsulConfig         *Consul.Config `json:"-"`
	ConsulServiceName    string         `json:"service_name" config:"APP_SERVICE_NAME"`
	ConsulServiceAddr    string         `json:"-" config:"APP_SERVICE_ADDR"`
	ConsulServicePort    string         `json:"service_port" config:"APP_SERVICE_PORT"`
}

func Init() (*AppConfig, error) {
	var zookeeperAddresses string
	var configPath string

	flag.StringVar(
		&zookeeperAddresses,
		"zookeeper-addrs",
		os.Getenv("APP_ZOOKEEPER_ADDRS"),
		"zookeeper addresses",
	)

	flag.StringVar(
		&configPath,
		"config-path",
		os.Getenv("APP_CONFIG_PATH"),
		"config path in zookeeper",
	)

	appConfig := &AppConfig{
		ClientEndpointAddr:   _DEFAULT_CLIENT_ENDPOINT_ADDRESS,
		ServiceEndpointAddr:  _DEFAULT_SERVICE_ENDPOINT_ADDRESS,
		RedisAddr:            _DEFAULT_REDIS_ADDRESS,
		BearerTokenDuration:  _DEFAULT_BEARER_TOKEN_DURATION,
		RefreshTokenDuration: _DEFAULT_REFRESH_TOKEN_DURATION,
		ConsulAddr:           _DEFAULT_CONSUL_ADDRESS,
		ConsulServiceName:    _DEFAULT_SERVICE_NAME,
		ConsulServicePort:    _DEFAULT_SERVICE_HEALTHY_PORT,
	}

	log.Println("Reading configuration from environment and default values")
	envConfig := Configurator.ReadConfigValues(appConfig)

	if zookeeperAddresses != "" {
		log.Printf("Connecting to Zookeeper at addresses: %s\n", zookeeperAddresses)
		zookeeperConfig, err := readConfigInZooKeeper(zookeeperAddresses, configPath)
		if err != nil {
			return nil, err
		}

		zkEnvConfig := Configurator.ReadConfigValuesFromStruct(zookeeperConfig)
		Configurator.ApplyConfigValues(appConfig, zkEnvConfig)
		log.Println("Configuration updated with values from Zookeeper")
		Configurator.Print(appConfig)
	}

	Configurator.ApplyConfigValues(appConfig, envConfig)
	log.Println("Configuration updated with values from environment variables")
	Configurator.Print(appConfig)

	Configurator.ReplaceDashWithEmpty(appConfig)
	log.Println("Replaced dashes with empty strings in configuration keys")
	Configurator.Print(appConfig)
	if appConfig.Id == "" {

		serviceId, err := generateId("Jwtissuer-")
		if err != nil {
			return nil, err
		}
		appConfig.Id = serviceId

	}

	appConfig.ConsulConfig = &Consul.Config{
		ServiceId:      appConfig.Id,
		ServiceName:    appConfig.ConsulServiceName,
		ServicePort:    appConfig.ConsulServicePort,
		ServiceAddress: appConfig.ConsulServiceAddr,
	}

	log.Println("Configuration initialization complete")
	return appConfig, nil
}

func readConfigInZooKeeper(zookeeperAddresses, configPath string) (*AppConfig, error) {
	log.Printf("Connecting to Zookeeper servers: %s\n", zookeeperAddresses)
	servers := strings.Split(zookeeperAddresses, ",")

	conn, _, err := zk.Connect(servers, 5*time.Second)
	if err != nil {
		log.Printf("Failed to connect to Zookeeper: %v\n", err)
		return nil, err
	}
	defer conn.Close()

	if configPath == "" {
		configPath = _DEFAULT_ZK_CONFIG_PATH
		log.Printf("Using default Zookeeper config path: %s\n", configPath)
	}

	log.Printf("Fetching configuration from Zookeeper path: %s\n", configPath)
	configBytes, _, err := conn.Get(configPath)
	if err != nil {
		if err == zk.ErrNoNode {
			log.Printf("Config path not found in Zookeeper, creating default configuration at %s\n", configPath)
			config := AppConfig{
				ClientEndpointAddr:   _DEFAULT_CLIENT_ENDPOINT_ADDRESS,
				ServiceEndpointAddr:  _DEFAULT_SERVICE_ENDPOINT_ADDRESS,
				RedisAddr:            _DEFAULT_REDIS_ADDRESS,
				BearerTokenDuration:  _DEFAULT_BEARER_TOKEN_DURATION,
				RefreshTokenDuration: _DEFAULT_REFRESH_TOKEN_DURATION,
				ConsulAddr:           _DEFAULT_CONSUL_ADDRESS,
				ConsulServiceName:    _DEFAULT_SERVICE_NAME,
				ConsulServicePort:    _DEFAULT_SERVICE_HEALTHY_PORT,
			}
			configBytes, err := json.Marshal(config)
			if err != nil {
				log.Printf("Error marshalling default config: %v\n", err)
				return nil, err
			}

			acl := zk.WorldACL(zk.PermAll)

			log.Printf("Creating Zookeeper path recursively for config at %s\n", configPath)
			if err := createPathRecursive(conn, configPath, configBytes, acl); err != nil {
				return nil, err
			}

			log.Println("Default configuration created in Zookeeper")
			return &config, nil
		}

		return nil, err
	}

	config := &AppConfig{}
	if err := json.Unmarshal(configBytes, config); err != nil {
		log.Printf("Error unmarshalling configuration: %v\n", err)
		return nil, err
	}

	log.Println("Successfully read configuration from Zookeeper")
	return config, nil
}

func createPathRecursive(conn *zk.Conn, path string, data []byte, acl []zk.ACL) error {
	log.Printf("Creating Zookeeper path recursively: %s\n", path)
	pathParts := strings.Split(path, "/")

	currentPath := ""
	for _, part := range pathParts {
		if part == "" {
			continue
		}
		currentPath += "/" + part

		exists, _, err := conn.Exists(currentPath)
		if err != nil {
			log.Printf("Error checking existence of path %s: %v\n", currentPath, err)
			return fmt.Errorf("error checking existence of path %s: %v", currentPath, err)
		}

		if !exists {
			log.Printf("Creating path %s\n", currentPath)
			_, err = conn.Create(currentPath, nil, 0, acl)
			if err != nil {
				log.Printf("Error creating path %s: %v\n", currentPath, err)
				return fmt.Errorf("error creating path %s: %v", currentPath, err)
			}
			log.Printf("Created path: %s\n", currentPath)
		}
	}

	log.Printf("Setting data for path %s\n", currentPath)
	_, err := conn.Set(currentPath, data, -1)
	if err != nil {
		log.Printf("Error setting data for path %s: %v\n", currentPath, err)
		return fmt.Errorf("error setting data for path %s: %v", currentPath, err)
	}

	log.Printf("Data at path %s has been updated to: %s\n", currentPath, string(data))
	return nil
}

func generateId(prefix string) (string, error) {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}
