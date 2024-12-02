package Kafkerclient

import (
	"context"
	"math/rand"
	ServiceKafker "peergrine/grpc/servicekafker"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const _MAX_RETRIES = 10
const _MIN_DELAY = 1 * time.Second
const _MAX_DELAY = 5 * time.Second

type Client struct {
	conn   *grpc.ClientConn
	client ServiceKafker.KafkerClient
}

func New(addr string) (*Client, error) {
	var err error

	for attempt := 1; attempt <= _MAX_RETRIES; attempt++ {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {

			client := ServiceKafker.NewKafkerClient(conn)
			return &Client{conn, client}, nil
		}

		delay := time.Duration(rand.Int63n(int64(_MAX_DELAY-_MIN_DELAY))) + _MIN_DELAY
		time.Sleep(delay)
	}

	return nil, err
}

func (c *Client) RequestPartition(serviceId string, serviceName string, topicName string) (int32, error) {
	req := &ServiceKafker.RequestKafkaPartitionReq{
		ServiceName: serviceName,
		TopicName:   topicName,
		ServiceId:   serviceId,
	}

	res, err := c.client.RequestKafkaPartition(context.Background(), req)

	if err != nil {
		return -1, err
	}

	return res.PartitionId, nil
}

func (c *Client) ReleasePartition(serviceId string) (string, error) {

	req := &ServiceKafker.ReleaseKafkaPartitionReq{
		ServiceId: serviceId,
	}

	res, err := c.client.ReleaseKafkaPartition(context.Background(), req)
	if err != nil {
		return "", err
	}

	return res.Message, nil
}
