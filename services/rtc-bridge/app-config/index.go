package appconfig

import (
	"flag"
	"log"
	"os"
	Configurator "peergrine/utils/configurator"

	"github.com/google/uuid"
)

const (
	_DEFAULT_ADDRESS           = ":80"
	_DEFAULT_AUTHORIZE_ADDRESS = "" // auth:50051
	_DEFAULT_REDIS_ADDRESS     = "" // redis:6379
	_DEFAULT_PULSAR_ADDRESSES  = "" // pulsar://pulsar-broker:6650
	_DEFAULT_PULSAR_TOPIC      = "RtcBridge"
	_DEFAULT_UNIFIED_MESSAGE   = "false"
	_DEFAULT_ZK_CONFIG_PATH    = "/rtc-bridge"
)

type AppConfig struct {
	Id                string `json:"-" config:"APP_ID"`
	Addr              string `json:"address" config:"APP_ADDR"`
	AuthAddr          string `json:"auth_address" config:"APP_AUTH_ADDR"`
	RedisAddr         string `json:"redis_address" config:"APP_REDIS_ADDR"`
	PulsarAddrs       string `json:"pulsar_addresses" config:"APP_PULSAR_ADDRS"`
	PulsarTopic       string `json:"pulsar_topic" config:"APP_PULSAR_TOPIC"`
	UnifiedMessageStr string `json:"unified_message" config:"APP_UNIFIED_MESSAGE"`
	UnifiedMessage    bool   `json:"-"`
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
		PulsarAddrs:       _DEFAULT_PULSAR_ADDRESSES,
		PulsarTopic:       _DEFAULT_PULSAR_TOPIC,
		UnifiedMessageStr: _DEFAULT_UNIFIED_MESSAGE,
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

	appConfig.UnifiedMessage = appConfig.UnifiedMessageStr == "true"

	log.Println("AppConfig initialized successfully.")
	return appConfig, nil
}
