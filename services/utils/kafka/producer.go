package kafka

import (
	"log"
	"math/rand"
	"time"

	"github.com/IBM/sarama"
)

const _PRODUCER_MAX_RETRIES = 10
const _PRODUCER_MIN_DELAY = 1 * time.Second
const _PRODUCER_MAX_DELAY = 5 * time.Second

type Producer struct {
	producer sarama.SyncProducer
}

func NewProducer(brokers []string) (*Producer, error) {
	var client sarama.SyncProducer
	var err error

	for attempt := 1; attempt <= _PRODUCER_MAX_RETRIES; attempt++ {
		client, err = sarama.NewSyncProducer(brokers, sarama.NewConfig())
		if err == nil {
			return &Producer{producer: client}, nil
		}

		log.Printf("Attempt %d/%d failed to connect to Kafka brokers: %v", attempt, _PRODUCER_MAX_RETRIES, err)

		delay := time.Duration(rand.Int63n(int64(_PRODUCER_MAX_DELAY-_PRODUCER_MIN_DELAY))) + _PRODUCER_MIN_DELAY
		log.Printf("Retrying in %v...", delay)
		time.Sleep(delay)
	}

	return nil, err
}

func (p *Producer) SendMessage(topic string, message []byte, partition int32) (int32, int64, error) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	if partition != -1 {
		msg.Partition = partition
	}

	return p.producer.SendMessage(msg)
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
