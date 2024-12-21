package appconfig

import (
	"flag"
	"log"
	"os"
	Configurator "peergrine/utils/configurator"
	Consul "peergrine/utils/consul/service"

	"github.com/google/uuid"
)

const (
	_DEFAULT_ADDRESS              = ":80"
	_DEFAULT_AUTHORIZE_ADDRESS    = "" //auth:50051
	_DEFAULT_REDIS_ADDRESS        = "" //redis:6379
	_DEFAULT_KAFKER_ADDRESS       = "" //kafker:50051
	_DEFAULT_KAFKA_ADDRESS        = "" //kafka:9092
	_DEFAULT_KAFKA_TOPIC          = "RtcBridge"
	_DEFAULT_UNIFIED_MESSAGE      = "false"
	_DEFAULT_CONSUL_ADDRESS       = "" //consul:8500
	_DEFAULT_SERVICE_NAME         = "RtcBridge"
	_DEFAULT_SERVICE_HEALTHY_PORT = "4000"
	_DEFAULT_ZK_CONFIG_PATH       = "/rtc-bridge"
)

type AppConfig struct {
	Id                string         `json:"-" config:"APP_ID"`
	Addr              string         `json:"address" config:"APP_ADDR"`
	AuthAddr          string         `json:"auth_address" config:"APP_AUTH_ADDR"`
	RedisAddr         string         `json:"redis_address" config:"APP_REDIS_ADDR"`
	KafkerAddr        string         `json:"kafker_address" config:"APP_KAFKER_ADDR"`
	KafkaAddr         string         `json:"kafka_address" config:"APP_KAFKA_ADDR"`
	KafkaTopic        string         `json:"kafka_topic" config:"APP_KAFKA_TOPIC"`
	UnifiedMessageStr string         `json:"unified_message" config:"APP_UNIFIED_MESSAGE"`
	UnifiedMessage    bool           `json:"-"`
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
		UnifiedMessageStr: _DEFAULT_UNIFIED_MESSAGE,
		ConsulAddr:        _DEFAULT_CONSUL_ADDRESS,
		ConsulServiceName: _DEFAULT_SERVICE_NAME,
		ConsulServicePort: _DEFAULT_SERVICE_HEALTHY_PORT,
	}

	log.Println("Reading environment configuration values...")
	envConfig := Configurator.ReadConfigValues(appConfig)

	if zookeeperAddresses != "" {
		log.Printf("Reading configuration from Zookeeper at %s...\n", zookeeperAddresses)

		if configPath == "" {
			configPath = _DEFAULT_ZK_CONFIG_PATH
			log.Printf("Using default Zookeeper config path: %s\n", configPath)
		}

		zkConfig, err := Configurator.ReadConfigFromZooKeeper(zookeeperAddresses, configPath, *appConfig)
		if err != nil {
			log.Printf("error reading config from Zookeeper: %v\n", err)
		} else {
			zkEnvConfig := Configurator.ReadConfigValuesFromStruct(zkConfig)
			log.Println("Applying Zookeeper configuration values...")
			Configurator.ApplyConfigValues(appConfig, zkEnvConfig)
			Configurator.Print(appConfig)
		}
	}

	log.Println("Applying environment configuration values...")
	Configurator.ApplyConfigValues(appConfig, envConfig)
	Configurator.Print(appConfig)

	log.Println("Replacing dashes in configuration keys with empty strings...")
	Configurator.ReplaceDashWithEmpty(appConfig)
	Configurator.Print(appConfig)

	if appConfig.Id == "" {
		appConfig.Id = "RtcBridge-" + uuid.New().String()
	}

	appConfig.ConsulConfig = &Consul.Config{
		ServiceId:      appConfig.Id,
		ServiceName:    appConfig.ConsulServiceName,
		ServiceAddress: appConfig.ConsulServiceAddr,
		ServicePort:    appConfig.ConsulServicePort,
	}

	appConfig.UnifiedMessage = appConfig.UnifiedMessageStr == "true"

	log.Println("AppConfig initialized successfully.")
	return appConfig, nil
}
