package Kafkerclient

import (
	"context"
	ServiceKafker "peergrine/grpc/servicekafker"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client ServiceKafker.KafkerClient
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := ServiceKafker.NewKafkerClient(conn)
	return &Client{conn, client}, nil
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
