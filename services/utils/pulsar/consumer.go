package pulsar

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

const _CONSUMER_MAX_RETRIES = 10
const _CONSUMER_MIN_DELAY = 1 * time.Second
const _CONSUMER_MAX_DELAY = 5 * time.Second

type Listeners struct {
	mux   *sync.RWMutex
	chans map[int]chan []byte
}

type Consumer struct {
	consumer  pulsar.Consumer
	listeners *Listeners
	stopFunc  context.CancelFunc
}

func newConsumer(client pulsar.Client, topic string, key string) (*Consumer, error) {
	var consumer pulsar.Consumer
	var err error

	options := pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: key,
		MessageChannel:   make(chan pulsar.ConsumerMessage, 100),
		KeySharedPolicy:  &pulsar.KeySharedPolicy{},
	}

	listeners := &Listeners{
		mux:   new(sync.RWMutex),
		chans: make(map[int]chan []byte),
	}

	for attempt := 1; attempt <= _CONSUMER_MAX_RETRIES; attempt++ {
		consumer, err = client.Subscribe(options)
		if err == nil {
			ctx, cancel := context.WithCancel(context.Background())

			go consumerListenMessages(ctx, consumer, key, listeners)

			return &Consumer{
				consumer:  consumer,
				listeners: listeners,
				stopFunc:  cancel,
			}, nil
		}

		log.Printf("Attempt %d/%d failed to connect to Pulsar brokers: %v", attempt, _CONSUMER_MAX_RETRIES, err)

		delay := time.Duration(rand.Int63n(int64(_CONSUMER_MAX_DELAY-_CONSUMER_MIN_DELAY))) + _CONSUMER_MIN_DELAY
		log.Printf("Retrying in %v...", delay)
		time.Sleep(delay)
	}

	return nil, err
}

func consumerListenMessages(ctx context.Context, consumer pulsar.Consumer, key string, listeners *Listeners) {
	for {
		select {
		case msg := <-consumer.Chan():
			if msg.Key() == key {
				listeners.mux.RLock()
				for _, ch := range listeners.chans {
					ch <- msg.Payload()
				}
				listeners.mux.RUnlock()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Consumer) ListenMessages(ctx context.Context, buf int) <-chan []byte {
	c.listeners.mux.Lock()
	defer c.listeners.mux.Unlock()

	ch := make(chan []byte)
	id := len(c.listeners.chans)

	c.listeners.chans[id] = ch
	return ch
}

func (c *Consumer) Close() {
	c.stopFunc()
	c.consumer.Close()
}
