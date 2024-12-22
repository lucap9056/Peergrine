package pulsar

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

const _PRODUCER_MAX_RETRIES = 10
const _PRODUCER_MIN_DELAY = 1 * time.Second
const _PRODUCER_MAX_DELAY = 5 * time.Second

type Producer struct {
	producer pulsar.Producer
}

func newProducer(client pulsar.Client, topic string) (*Producer, error) {
	var producer pulsar.Producer
	var err error

	options := pulsar.ProducerOptions{
		Topic: topic,
	}

	for attempt := 1; attempt <= _PRODUCER_MAX_RETRIES; attempt++ {
		producer, err = client.CreateProducer(options)
		if err == nil {
			return &Producer{producer: producer}, nil
		}

		log.Printf("Attempt %d/%d failed to connect to Pulsar brokers: %v", attempt, _PRODUCER_MAX_RETRIES, err)

		delay := time.Duration(rand.Int63n(int64(_PRODUCER_MAX_DELAY-_PRODUCER_MIN_DELAY))) + _PRODUCER_MIN_DELAY
		log.Printf("Retrying in %v...", delay)
		time.Sleep(delay)
	}

	return nil, err
}

func (p *Producer) SendMessage(key string, message []byte) (pulsar.MessageID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	return p.producer.Send(ctx, &pulsar.ProducerMessage{
		Payload: message,
		Key:     key,
	})
}

func (p *Producer) Close() {
	p.producer.Close()
}
