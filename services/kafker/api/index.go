package kafkerapi

import (
	"context"
	"fmt"
	"log"
	"net"
	gRPCServiceKafker "peergrine/grpc/servicekafker"
	Storage "peergrine/kafker/storage"
	"time"

	"google.golang.org/grpc"
)

type KafkerServiceServer struct {
	gRPCServiceKafker.UnimplementedKafkerServer
	storage           *Storage.Storage
	healthCheckTicker *time.Ticker
}

// RequestKafkaPartition handles the request for a Kafka partition for a specific service.
func (k *KafkerServiceServer) RequestKafkaPartition(ctx context.Context, req *gRPCServiceKafker.RequestKafkaPartitionReq) (*gRPCServiceKafker.RequestKafkaPartitionRes, error) {
	log.Printf("Requesting Kafka partition for service: %s on topic: %s", req.ServiceId, req.TopicName)

	partitionId, err := k.storage.RequestKafkaPartition(req.TopicName, req.ServiceName, req.ServiceId)
	if err != nil {
		log.Printf("Error requesting Kafka partition for service: %s on topic: %s, error: %v", req.ServiceId, req.TopicName, err)
		return nil, fmt.Errorf("failed to request Kafka partition for service %s on topic %s: %w", req.ServiceName, req.TopicName, err)
	}

	log.Printf("Successfully assigned partition %d to service: %s on topic: %s", partitionId, req.ServiceId, req.TopicName)
	return &gRPCServiceKafker.RequestKafkaPartitionRes{
		PartitionId: int32(partitionId),
	}, nil
}

// ReleaseKafkaPartition handles the request to release a Kafka partition for a specific service.
func (k *KafkerServiceServer) ReleaseKafkaPartition(ctx context.Context, req *gRPCServiceKafker.ReleaseKafkaPartitionReq) (*gRPCServiceKafker.ReleaseKafkaPartitionRes, error) {
	log.Printf("Releasing Kafka partition for service: %s", req.ServiceId)

	err := k.storage.ReleaseKafkaPartition(req.ServiceId)
	if err != nil {
		log.Printf("Error releasing Kafka partition for service: %s, error: %v", req.ServiceId, err)
		return nil, fmt.Errorf("failed to release Kafka partition for service %s: %w", req.ServiceId, err)
	}

	log.Printf("Successfully released Kafka partition for service: %s", req.ServiceId)
	return &gRPCServiceKafker.ReleaseKafkaPartitionRes{
		Message: "Success",
	}, nil
}

type Server struct {
	*grpc.Server
	service *KafkerServiceServer
}

// New initializes a new Kafker service server with the provided storage manager.
func New(storage *Storage.Storage) (*Server, error) {
	server := grpc.NewServer()
	kafkerServer := &KafkerServiceServer{
		storage: storage,
	}

	gRPCServiceKafker.RegisterKafkerServer(server, kafkerServer)
	return &Server{server, kafkerServer}, nil
}

// Run starts the gRPC server on the specified address.
func (s *Server) Run(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start listener on address %s: %w", addr, err)
	}

	return s.Serve(listener)
}

// Close stops the health check ticker if it's running.
func (s *Server) Close() {
	if s.service.healthCheckTicker != nil {
		s.service.healthCheckTicker.Stop()
	}
}
