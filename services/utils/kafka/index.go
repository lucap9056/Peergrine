package kafka

import "strings"

type Client struct {
	*Consumer
	*Producer
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

func New(brokers string) (*Client, error) {
	brokersArray := strings.Split(brokers, ",")

	consumer, err := NewConsumer(brokersArray)
	if err != nil {
		return nil, err
	}

	producer, err := NewProducer(brokersArray)
	if err != nil {
		return nil, err
	}

	client := &Client{
		consumer,
		producer,
	}

	return client, nil
}
