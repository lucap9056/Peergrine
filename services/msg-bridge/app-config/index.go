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
	_DEFAULT_ADDRESS              = ":80"
	_DEFAULT_AUTHORIZE_ADDRESS    = "" //auth:50051
	_DEFAULT_REDIS_ADDRESS        = "" //redis:6379
	_DEFAULT_KAFKER_ADDRESS       = "" //kafker:50051
	_DEFAULT_KAFKA_ADDRESS        = "" //kafka:9092
	_DEFAULT_KAFKA_TOPIC          = "MsgBridge"
	_DEFAULT_CONSUL_ADDRESS       = "" //consul:8500
	_DEFAULT_SERVICE_NAME         = "MsgBridge"
	_DEFAULT_SERVICE_HEALTHY_PORT = "4000"
	_DEFAULT_ZK_CONFIG_PATH       = "/msg-bridge"
)

type AppConfig struct {
	Id                string         `json:"-" config:"APP_ID"`
	Addr              string         `json:"address" config:"APP_ADDR"`
	AuthAddr          string         `json:"auth_address" config:"APP_AUTH_ADDR"`
	RedisAddr         string         `json:"redis_address" config:"APP_REDIS_ADDR"`
	KafkaAddr         string         `json:"kafka_address" config:"APP_KAFKA_ADDR"`
	KafkerAddr        string         `json:"kafker_address" config:"APP_KAFKER_ADDR"`
	KafkaTopic        string         `json:"-"`
	ConsulAddr        string         `json:"consul_address" config:"APP_CONSUL_ADDR"`
	ConsulConfig      *Consul.Config `json:"-"`
	ConsulServiceName string         `json:"service_name" config:"APP_SERVICE_NAME"`
	ConsulServiceAddr string         `json:"-" config:"APP_SERVICE_ADDR"`
	ConsulServicePort string         `json:"service_port" config:"APP_SERVICE_PORT"`
}

func Init() (*AppConfig, error) {
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
		AuthAddr:          _DEFAULT_AUTHORIZE_ADDRESS,
		RedisAddr:         _DEFAULT_REDIS_ADDRESS,
		KafkerAddr:        _DEFAULT_KAFKER_ADDRESS,
		KafkaAddr:         _DEFAULT_KAFKA_ADDRESS,
		KafkaTopic:        _DEFAULT_KAFKA_TOPIC,
		ConsulAddr:        _DEFAULT_CONSUL_ADDRESS,
		ConsulServiceName: _DEFAULT_SERVICE_NAME,
		ConsulServicePort: _DEFAULT_SERVICE_HEALTHY_PORT,
	}

	log.Println("Reading environment configuration values...")
	envConfig := Configurator.ReadConfigValues(appConfig)

	if zookeeperAddresses != "" {
		log.Printf("Connecting to Zookeeper at %s...\n", zookeeperAddresses)
		zkConfig, err := readConfigInZooKeeper(zookeeperAddresses, configPath)
		if err != nil {
			return nil, fmt.Errorf("error reading config from Zookeeper: %w", err)
		}

		zkEnvConfig := Configurator.ReadConfigValuesFromStruct(zkConfig)
		log.Println("Applying Zookeeper configuration values...")
		Configurator.ApplyConfigValues(appConfig, zkEnvConfig)
		Configurator.Print(appConfig)
	}

	log.Println("Applying environment configuration values...")
	Configurator.ApplyConfigValues(appConfig, envConfig)
	Configurator.Print(appConfig)

	log.Println("Replacing dashes with empty strings in configuration keys...")
	Configurator.ReplaceDashWithEmpty(appConfig)
	Configurator.Print(appConfig)

	if appConfig.Id == "" {
		appConfig.Id = "MsgBridge-" + uuid.New().String()
	}

	appConfig.ConsulConfig = &Consul.Config{
		ServiceId:      appConfig.Id,
		ServiceName:    appConfig.ConsulServiceName,
		ServiceAddress: appConfig.ConsulServiceAddr,
		ServicePort:    appConfig.ConsulServicePort,
	}

	log.Println("Configuration initialization complete.")
	return appConfig, nil
}

func readConfigInZooKeeper(zookeeperAddresses, configPath string) (*AppConfig, error) {
	servers := strings.Split(zookeeperAddresses, ",")

	log.Println("Connecting to Zookeeper servers...")
	conn, _, err := zk.Connect(servers, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Zookeeper: %w", err)
	}
	defer conn.Close()

	if configPath == "" {
		configPath = _DEFAULT_ZK_CONFIG_PATH
		log.Printf("Using default Zookeeper config path: %s\n", configPath)
	}

	log.Printf("Reading configuration from Zookeeper path: %s\n", configPath)
	configBytes, _, err := conn.Get(configPath)
	if err != nil {
		if err == zk.ErrNoNode {
			log.Printf("Config path %s not found. Creating default configuration...\n", configPath)
			config := AppConfig{
				Addr:              _DEFAULT_ADDRESS,
				AuthAddr:          _DEFAULT_AUTHORIZE_ADDRESS,
				RedisAddr:         _DEFAULT_REDIS_ADDRESS,
				KafkerAddr:        _DEFAULT_KAFKER_ADDRESS,
				KafkaAddr:         _DEFAULT_KAFKA_ADDRESS,
				ConsulAddr:        _DEFAULT_CONSUL_ADDRESS,
				ConsulServiceName: _DEFAULT_SERVICE_NAME,
				ConsulServicePort: _DEFAULT_SERVICE_HEALTHY_PORT,
			}
			configBytes, err := json.Marshal(config)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal default config: %w", err)
			}

			acl := zk.WorldACL(zk.PermAll)

			if err := createPathRecursive(conn, configPath, configBytes, acl); err != nil {
				return nil, fmt.Errorf("error creating Zookeeper config path: %w", err)
			}

			log.Println("Default configuration created in Zookeeper.")
			return &config, nil
		}
		return nil, fmt.Errorf("error reading config from Zookeeper: %w", err)
	}

	config := &AppConfig{}
	if err := json.Unmarshal(configBytes, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Zookeeper config: %w", err)
	}

	log.Println("Successfully loaded configuration from Zookeeper.")
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

	log.Printf("Data at path %s updated: %s\n", currentPath, string(data))
	return nil
}
