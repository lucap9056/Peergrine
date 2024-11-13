package kafka

import (
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

type Client struct {
	brokers []string
	client  sarama.Client
}

type KafkaConfig struct {
	Features                    map[string]interface{} `json:"features"`
	ListenerSecurityProtocolMap map[string]string      `json:"listener_security_protocol_map"`
	Endpoints                   []string               `json:"endpoints"`
	JmxPort                     int                    `json:"jmx_port"`
	Port                        int                    `json:"port"`
	Host                        string                 `json:"host"`
	Version                     int                    `json:"version"`
	Timestamp                   string                 `json:"timestamp"`
}

// New creates a new Kafka Client and establishes a connection to one of the provided brokers.
func New(brokersAddress string) (*Client, error) {

	brokers := strings.Split(brokersAddress, ",")

	config := sarama.NewConfig()
	config.Net.DialTimeout = 10 * time.Second
	config.Net.ReadTimeout = 10 * time.Second
	config.Net.WriteTimeout = 10 * time.Second

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return &Client{
		brokers: brokers,
		client:  client,
	}, nil
}

// ReadPartitions retrieves the partitions for a given topic from the Kafka broker.
func (kafka *Client) ReadPartitions(topic string) ([]int32, error) {

	partitions, err := kafka.client.Partitions(topic)
	if err != nil {
		return nil, fmt.Errorf("error reading partitions for topic %s: %w", topic, err)
	}

	return partitions, nil
}

// Close gracefully closes the connection to the Kafka broker.
func (kafka *Client) Close() error {
	return kafka.client.Close()
}
