package appconfig

import (
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
	"github.com/google/uuid"
)

const (
	_DEFAULT_ADDRESS              = ":50051"
	_DEFAULT_CLUSTER_MODE         = "false"
	_DEFAULT_KAFKA_ADDRESS        = "kafka:9092"
	_DEFAULT_CONSUL_ADDRESS       = "consul:8500"
	_DEFAULT_SERVICE_NAME         = "Kafker"
	_DEFAULT_SERVICE_HEALTHY_PORT = "4000"
	_DEFAULT_ZK_CONFIG_PATH       = "/kafker"
)

type AppConfig struct {
	Id                string         `json:"-" config:"APP_ID"`
	Addr              string         `json:"address" config:"APP_ADDR"`
	ClusterMode       bool           `json:"-"`
	ClusterModeStr    string         `json:"cluster_mode" config:"APP_CLUSTER_MODE"`
	KafkaAddr         string         `json:"kafka_address" config:"APP_KAFKA_ADDR"`
	ConsulAddr        string         `json:"consul_address" config:"APP_CONSUL_ADDR"`
	ConsulConfig      *Consul.Config `json:"-"`
	ConsulServiceName string         `json:"consul_service_name" config:"APP_SERVICE_NAME"`
	ConsulServiceAddr string         `json:"-" config:"APP_SERVICE_ADDR"`
	ConsulServicePort string         `json:"consul_service_port" config:"APP_SERVICE_PORT"`
}

func Init() (*AppConfig, *zk.Conn, error) {

	var zookeeperAddresses string
	var configPath string

	flag.StringVar(
		&zookeeperAddresses,
		"ZOOKEEPER_ADDRS",
		os.Getenv("APP_ZOOKEEPER_ADDRS"),
		"zookeeper addresses",
	)

	flag.StringVar(
		&configPath,
		"CONFIG_PATH",
		os.Getenv("CONFIG_PATH"),
		"config path in zookeeper",
	)

	appConfig := &AppConfig{
		Addr:              _DEFAULT_ADDRESS,
		ClusterModeStr:    _DEFAULT_CLUSTER_MODE,
		ConsulAddr:        _DEFAULT_CONSUL_ADDRESS,
		ConsulServiceName: _DEFAULT_SERVICE_NAME,
		ConsulServicePort: _DEFAULT_SERVICE_HEALTHY_PORT,
	}

	log.Println("Reading configuration values from environment and default settings")
	envConfig := Configurator.ReadConfigValues(appConfig)

	if zookeeperAddresses != "" {
		log.Printf("Connecting to Zookeeper at addresses: %s\n", zookeeperAddresses)
		zookeeperConfig, conn, err := readConfigInZooKeeper(zookeeperAddresses, configPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read configuration from Zookeeper: %w", err)
		}

		zkEnvConfig := Configurator.ReadConfigValuesFromStruct(zookeeperConfig)
		Configurator.ApplyConfigValues(appConfig, zkEnvConfig)
		log.Println("Loaded configuration from Zookeeper")
		Configurator.Print(appConfig)

		Configurator.ApplyConfigValues(appConfig, envConfig)
		log.Println("Loaded additional configuration from environment variables")
		Configurator.Print(appConfig)

		Configurator.ReplaceDashWithEmpty(appConfig)
		log.Println("Completed dash replacements in configuration keys")
		Configurator.Print(appConfig)

		if appConfig.Id == "" {
			appConfig.Id = "Kafker-" + uuid.New().String()
		}

		appConfig.ClusterMode = appConfig.ClusterModeStr == "true"
		appConfig.ConsulConfig = &Consul.Config{
			ServiceId:      appConfig.Id,
			ServiceName:    appConfig.ConsulServiceName,
			ServiceAddress: appConfig.ConsulServiceAddr,
			ServicePort:    appConfig.ConsulServicePort,
		}

		log.Println("Initialization complete with Zookeeper configuration")
		return appConfig, conn, nil

	} else {
		Configurator.ApplyConfigValues(appConfig, envConfig)
		log.Println("Loaded configuration from environment variables without Zookeeper")
		Configurator.Print(appConfig)

		Configurator.ReplaceDashWithEmpty(appConfig)
		log.Println("Completed dash replacements in configuration keys")
		Configurator.Print(appConfig)

		if appConfig.Id == "" {
			appConfig.Id = "Kafker-" + uuid.New().String()
		}
	}

	log.Println("Initialization complete without Zookeeper")
	return appConfig, nil, nil
}

func readConfigInZooKeeper(zookeeperAddresses, configPath string) (*AppConfig, *zk.Conn, error) {
	log.Printf("Connecting to Zookeeper servers: %s\n", zookeeperAddresses)
	servers := strings.Split(zookeeperAddresses, ",")

	conn, _, err := zk.Connect(servers, 5*time.Second)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to connect to Zookeeper: %w", err)
	}

	if configPath == "" {
		configPath = _DEFAULT_ZK_CONFIG_PATH
		log.Printf("Using default Zookeeper config path: %s\n", configPath)
	}

	log.Printf("Reading configuration from Zookeeper path: %s\n", configPath)
	configBytes, _, err := conn.Get(configPath)
	if err != nil {
		if err == zk.ErrNoNode {
			log.Printf("Configuration path %s not found in Zookeeper, creating default configuration\n", configPath)
			return createDefaultZookeeperConfig(conn, configPath)
		}
		return nil, nil, fmt.Errorf("error reading configuration from Zookeeper: %w", err)
	}

	config := &AppConfig{}
	if err := json.Unmarshal(configBytes, config); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal configuration from Zookeeper: %w", err)
	}

	log.Println("Successfully read configuration from Zookeeper")
	return config, conn, nil
}

func createDefaultZookeeperConfig(conn *zk.Conn, configPath string) (*AppConfig, *zk.Conn, error) {
	config := &AppConfig{
		Addr:              _DEFAULT_ADDRESS,
		ClusterModeStr:    _DEFAULT_CLUSTER_MODE,
		ConsulAddr:        _DEFAULT_CONSUL_ADDRESS,
		ConsulServiceName: _DEFAULT_SERVICE_NAME,
		ConsulServicePort: _DEFAULT_SERVICE_HEALTHY_PORT,
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal default config: %w", err)
	}

	acl := zk.WorldACL(zk.PermAll)
	if err := createPathRecursive(conn, configPath, configBytes, acl); err != nil {
		return nil, nil, fmt.Errorf("failed to create Zookeeper path recursively: %w", err)
	}

	log.Println("Default configuration created in Zookeeper")
	return config, conn, nil
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
			return fmt.Errorf("error checking existence of path %s: %w", currentPath, err)
		}

		if !exists {
			_, err = conn.Create(currentPath, nil, 0, acl)
			if err != nil {
				return fmt.Errorf("error creating path %s: %w", currentPath, err)
			}
			log.Printf("Created path: %s\n", currentPath)
		}
	}

	_, err := conn.Set(currentPath, data, -1)
	if err != nil {
		return fmt.Errorf("error setting data for path %s: %w", currentPath, err)
	}

	log.Printf("Data at path %s has been updated to: %s\n", currentPath, string(data))
	return nil
}
