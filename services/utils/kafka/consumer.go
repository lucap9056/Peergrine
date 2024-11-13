package kafka

import (
	"log"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumer sarama.Consumer
}

func NewConsumer(brokers []string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	client, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{consumer: client}, nil
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
			handler <- msg.Value
			<-done
		case err := <-pc.Errors():
			log.Println("Error:", err)
		}
	}
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
