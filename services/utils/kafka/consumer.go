package kafka

import (
	"log"
	"math/rand"
	"time"

	"github.com/IBM/sarama"
)

const _CONSUMER_MAX_RETRIES = 10
const _CONSUMER_MIN_DELAY = 1 * time.Second
const _CONSUMER_MAX_DELAY = 5 * time.Second

type Consumer struct {
	consumer sarama.Consumer
}

func NewConsumer(brokers []string) (*Consumer, error) {
	var client sarama.Consumer
	var err error

	for attempt := 1; attempt <= _CONSUMER_MAX_RETRIES; attempt++ {
		client, err = sarama.NewConsumer(brokers, sarama.NewConfig())
		if err == nil {
			return &Consumer{consumer: client}, nil
		}

		log.Printf("Attempt %d/%d failed to connect to Kafka brokers: %v", attempt, _CONSUMER_MAX_RETRIES, err)

		delay := time.Duration(rand.Int63n(int64(_CONSUMER_MAX_DELAY-_CONSUMER_MIN_DELAY))) + _CONSUMER_MIN_DELAY
		log.Printf("Retrying in %v...", delay)
		time.Sleep(delay)
	}

	return nil, err
}

func (c *Consumer) ConsumeMessages(topic string, partition int32, offset int64, handler chan []byte, done chan interface{}) error {
	pc, err := c.consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		return err
	}
	defer pc.Close()

	for {
		select {
		case msg := <-pc.Messages():
			select {
			case handler <- msg.Value:
			case <-time.After(5 * time.Second):
				log.Println("Handler channel timeout")
			}
		case err := <-pc.Errors():
			log.Println("Error:", err)
		case <-done:
			log.Println("Stopping consumer")
			return nil
		}
	}
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
