package kafka

import "strings"

type Client struct {
	*Consumer
	*Producer
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
