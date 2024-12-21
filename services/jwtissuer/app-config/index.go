package appconfig

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"log"
	"os"
	Configurator "peergrine/utils/configurator"
	Consul "peergrine/utils/consul/service"
)

const (
	_DEFAULT_CLIENT_ENDPOINT_ADDRESS  = ":80"
	_DEFAULT_SERVICE_ENDPOINT_ADDRESS = ":50051"
	_DEFAULT_REDIS_ADDRESS            = "" //redis:6379
	_DEFAULT_BEARER_TOKEN_DURATION    = "3600"
	_DEFAULT_REFRESH_TOKEN_DURATION   = "7200"
	_DEFAULT_KAFKER_ADDRESS           = "" //kafker:50051
	_DEFAULT_KAFKA_ADDRESS            = "" //kafka:9092
	_DEFAULT_KAFKA_TOPIC              = "JwtIssuer"
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
	KafkaAddr            string         `json:"kafka_address" config:"APP_KAFKA_ADDR"`
	KafkerAddr           string         `json:"kafker_address" config:"APP_KAFKER_ADDR"`
	KafkaTopic           string         `json:"kafka_topic" config:"APP_KAFKA_TOPIC"`
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
		KafkerAddr:           _DEFAULT_KAFKER_ADDRESS,
		KafkaAddr:            _DEFAULT_KAFKA_ADDRESS,
		KafkaTopic:           _DEFAULT_KAFKA_TOPIC,
		ConsulAddr:           _DEFAULT_CONSUL_ADDRESS,
		ConsulServiceName:    _DEFAULT_SERVICE_NAME,
		ConsulServicePort:    _DEFAULT_SERVICE_HEALTHY_PORT,
	}

	log.Println("Reading configuration from environment and default values")
	envConfig := Configurator.ReadConfigValues(appConfig)

	if zookeeperAddresses != "" {
		log.Printf("Connecting to Zookeeper at addresses: %s\n", zookeeperAddresses)

		if configPath == "" {
			configPath = _DEFAULT_ZK_CONFIG_PATH
			log.Printf("Using default Zookeeper config path: %s\n", configPath)
		}

		zkConfig, err := Configurator.ReadConfigFromZooKeeper(zookeeperAddresses, configPath, *appConfig)
		if err != nil {
			log.Printf("error reading config from Zookeeper: %v\n", err)
		} else {
			zkEnvConfig := Configurator.ReadConfigValuesFromStruct(zkConfig)
			Configurator.ApplyConfigValues(appConfig, zkEnvConfig)
			log.Println("Configuration updated with values from Zookeeper")
			Configurator.Print(appConfig)
		}
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

func generateId(prefix string) (string, error) {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}
