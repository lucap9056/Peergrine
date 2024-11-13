package storage

import (
	"context"
	"fmt"
	"log"
	Kafka "peergrine/kafker/kafka"
	Consul "peergrine/utils/consul/manager"
	"strconv"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
)

type Service struct {
	Id        string
	Name      string
	Topic     string
	Partition string
}

type Topic struct {
	Mux                   *sync.Mutex         `json:"-"`
	MaximumPartitionCount int                 `json:"maximum_partition_count"`
	Services              map[string]*Service `json:"-"`
}

type Storage struct {
	zk     *zkStorage
	local  *localStorage
	kafka  *Kafka.Client
	consul *Consul.Manager
}

// New creates a new Storage instance.
func New(zookeeper *zk.Conn, kafka *Kafka.Client, consul *Consul.Manager, clusterMode bool) (*Storage, error) {
	storage := &Storage{
		local:  NewLocalStorage(),
		kafka:  kafka,
		consul: consul,
	}

	if clusterMode {
		zkStorage, err := NewZkStorage(zookeeper)
		if err != nil {
			return nil, err
		}
		storage.zk = zkStorage
	}

	return storage, nil
}
func (s *Storage) RequestKafkaPartition(topicName, serviceName, serviceId string) (int, error) {
	service := &Service{
		Id:    serviceId,
		Name:  serviceName,
		Topic: topicName,
	}

	if s.zk != nil {
		// Acquire write lock for ZooKeeper
		writeLock, err := s.zk.WLock(zkBasePath)
		if err != nil {
			return -1, fmt.Errorf("unable to acquire write lock for ZooKeeper: %w", err)
		}
		defer writeLock.WUnlock()
		log.Printf("Write lock acquired for ZooKeeper path %s\n", zkBasePath)

		topic, err := s.zk.GetTopic(topicName)
		if err != nil {
			return -1, fmt.Errorf("unable to retrieve topic %s from ZooKeeper: %w", topicName, err)
		}

		if topic == nil {
			partitions, err := s.kafka.ReadPartitions(topicName)
			if err != nil {
				return -1, fmt.Errorf("unable to retrieve partitions for topic %s after reconnect: %w", topicName, err)

			}

			partitionCount := len(partitions)
			log.Printf("Topic %s has %d partitions\n", topicName, partitionCount)

			if err := s.zk.AppendTopic(topicName, partitionCount); err != nil {
				return -1, fmt.Errorf("unable to append topic %s to ZooKeeper: %w", topicName, err)
			}
			log.Printf("Topic %s appended to ZooKeeper with %d partitions\n", topicName, partitionCount)

			service.Partition = "0"
			log.Printf("Service %s assigned to partition 0\n", serviceName)
			if err := s.zk.AppendService(service); err != nil {
				return -1, fmt.Errorf("unable to append service %s to partition 0: %w", serviceName, err)
			}
			log.Printf("Service %s successfully appended to partition 0 of topic %s\n", serviceName, topicName)

			return 0, nil
		} else {
			partitions, err := s.zk.GetTopicServicePartitions(topicName)
			if err != nil {
				return -1, fmt.Errorf("unable to retrieve partitions for topic %s: %w", topicName, err)
			}

			log.Printf("Topic %s has %d partitions\n", topicName, topic.MaximumPartitionCount)
			log.Printf("Existing partitions for topic %s: %v\n", topicName, partitions)
			partitionsMap := make(map[string]struct{})
			for _, partition := range partitions {
				partitionsMap[partition] = struct{}{}
			}

			for i := 0; i < topic.MaximumPartitionCount; i++ {
				partition := strconv.Itoa(i)
				if _, exists := partitionsMap[partition]; !exists {
					service.Partition = partition
					log.Printf("Service %s assigned to partition %s\n", serviceName, partition)
					if err := s.zk.AppendService(service); err != nil {
						return i, fmt.Errorf("unable to append service %s to partition %s: %w", serviceName, partition, err)
					}
					log.Printf("Service %s successfully appended to partition %s of topic %s\n", serviceName, partition, topicName)
					return i, nil
				}
			}
		}
	} else {
		s.local.Mux.Lock()
		defer s.local.Mux.Unlock()

		localTopic, exists := s.local.Topics[topicName]
		if !exists {
			partitions, err := s.kafka.ReadPartitions(topicName)
			if err != nil {
				return -1, fmt.Errorf("unable to retrieve partitions for topic %s after reconnect: %w", topicName, err)
			}

			localTopic = &Topic{
				Mux:                   new(sync.Mutex),
				MaximumPartitionCount: len(partitions),
				Services:              make(map[string]*Service),
			}

			s.local.Topics[topicName] = localTopic
			localTopic.Services["0"] = service
			s.local.Services[serviceId] = service
			log.Printf("Created local topic %s and assigned service %s to partition 0\n", topicName, serviceName)

			return 0, nil
		} else {
			for i := 0; i < localTopic.MaximumPartitionCount; i++ {
				partition := strconv.Itoa(i)
				if _, exists := localTopic.Services[partition]; !exists {
					localTopic.Services[partition] = service
					s.local.Services[serviceId] = service
					log.Printf("Assigned service %s to partition %s of local topic %s\n", serviceName, partition, topicName)
					return i, nil
				}
			}
		}
	}

	return 0, fmt.Errorf("no available partitions for topic %s", topicName)
}

// ReleaseKafkaPartition releases a Kafka partition for the given service ID.
func (s *Storage) ReleaseKafkaPartition(serviceId string) error {
	if s.zk != nil {
		if err := s.zk.RemoveService(serviceId); err != nil {
			return fmt.Errorf("failed to remove service %s from zk: %w", serviceId, err)
		}
		log.Printf("Removed service %s from zk\n", serviceId)
	} else {
		s.local.Mux.Lock()
		defer s.local.Mux.Unlock()

		service, exists := s.local.Services[serviceId]
		if !exists {
			return fmt.Errorf("service %s not found in local services", serviceId)
		}

		topic, exists := s.local.Topics[service.Topic]
		if !exists {
			return fmt.Errorf("topic %s not found for service %s", service.Topic, serviceId)
		}

		delete(s.local.Services, serviceId)
		delete(topic.Services, service.Partition)
		log.Printf("Removed service %s from topic %s at partition %s\n", serviceId, service.Topic, service.Partition)
	}

	return nil
}

// StartServiceHealthCheck starts the health check for services.
func (s *Storage) StartServiceHealthCheck(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping service health check...")
			return
		default:
			if s.zk != nil {
				if err := s.zkCheckAndRemoveUnhealthyServices(); err != nil {
					log.Println("Error while checking and removing unhealthy services:", err)
				}
			} else {
				if err := s.localCheckAndRemoveUnhealthyServices(); err != nil {
					log.Println("Error while checking and removing unhealthy services:", err)
				}
			}

			time.Sleep(60 * time.Second)
		}
	}
}

// zkCheckAndRemoveUnhealthyServices checks and removes unhealthy services in zk.
func (s *Storage) zkCheckAndRemoveUnhealthyServices() error {

	existedServices, serviceNames, err := func() (map[string]struct{}, []string, error) {

		rlock, err := s.zk.RLock(zkBasePath)
		if err != nil {
			return nil, nil, err
		}
		defer rlock.RUnlock()

		services, err := s.zk.GetAllServices()
		if err != nil {
			return nil, nil, err
		}

		existedServices := make(map[string]struct{})

		for _, service := range services {
			existedServices[service] = struct{}{}
		}

		serviceNames, err := s.zk.GetServiceNames()
		if err != nil {
			return nil, nil, err
		}

		return existedServices, serviceNames, nil
	}()

	if err != nil {
		return err
	}

	for _, serviceName := range serviceNames {
		services, err := s.consul.GetServices(serviceName)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, service := range services {
			delete(existedServices, service.ServiceID)
		}
	}

	wlock, err := s.zk.WLock(zkBasePath)
	if err != nil {
		return err
	}
	defer wlock.WUnlock()

	for service := range existedServices {
		err := s.zk.RemoveService(service)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

// localCheckAndRemoveUnhealthyServices checks and removes unhealthy services locally.
func (s *Storage) localCheckAndRemoveUnhealthyServices() error {
	s.local.Mux.Lock()
	defer s.local.Mux.Unlock()

	existedServices := make(map[string]*Service)
	serviceNames := make(map[string]struct{})

	for id, service := range s.local.Services {
		existedServices[id] = service
		serviceNames[service.Name] = struct{}{}
	}

	for serviceName := range serviceNames {
		services, err := s.consul.GetServices(serviceName)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, service := range services {
			delete(existedServices, service.ServiceID)
		}
	}

	for _, service := range existedServices {
		delete(s.local.Services, service.Id)

		topic, exists := s.local.Topics[service.Topic]
		if exists {
			delete(topic.Services, service.Partition)
		}
	}

	return nil
}
