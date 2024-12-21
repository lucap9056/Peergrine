package configurator

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

func ReadConfigFromZooKeeper[T any](zookeeperAddresses, configPath string, defaultConfig T) (*T, error) {
	servers := strings.Split(zookeeperAddresses, ",")

	log.Println("Connecting to Zookeeper servers...")
	conn, _, err := zk.Connect(servers, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to Zookeeper: %w", err)
	}
	defer conn.Close()

	log.Printf("Reading configuration from Zookeeper path: %s\n", configPath)
	configBytes, _, err := conn.Get(configPath)
	if err != nil {
		if err == zk.ErrNoNode {
			log.Printf("Config path %s not found. Creating default configuration...\n", configPath)

			configBytes, err := json.Marshal(defaultConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal default config: %w", err)
			}

			acl := zk.WorldACL(zk.PermAll)

			if err := createPathRecursive(conn, configPath, configBytes, acl); err != nil {
				return nil, fmt.Errorf("error creating Zookeeper config path: %w", err)
			}

			log.Println("Default configuration created in Zookeeper.")
			return &defaultConfig, nil
		}
		return nil, fmt.Errorf("error reading config from Zookeeper: %w", err)
	}

	var config T
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Zookeeper config: %w", err)
	}

	log.Println("Successfully loaded configuration from Zookeeper.")
	return &config, nil
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
