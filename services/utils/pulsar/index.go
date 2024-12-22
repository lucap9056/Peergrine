package pulsar

import (
	"context"

	"github.com/apache/pulsar-client-go/pulsar"
)

type Client struct {
	client   pulsar.Client
	consumer *Consumer
	producer *Producer
}

func New(brokers string, topic string, key string) (*Client, error) {

	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: brokers,
	})

	if err != nil {
		return nil, err
	}

	producer, err := newProducer(client, topic)

	if err != nil {
		return nil, err
	}

	consumer, err := newConsumer(client, topic, key)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:   client,
		producer: producer,
		consumer: consumer,
	}, nil
}

func (c *Client) SendMessage(key string, message []byte) (pulsar.MessageID, error) {
	return c.producer.SendMessage(key, message)
}

func (c *Client) ListenMessages(ctx context.Context, buf int) <-chan []byte {
	return c.consumer.ListenMessages(ctx, buf)
}

func (c *Client) Close() {
	c.producer.Close()
	c.consumer.Close()
}
