package storage

import (
	"errors"
	"log"
	"sync"
)

type localStorage struct {
	Mux      *sync.RWMutex
	Services map[string]*Service
	Topics   map[string]*Topic
}

// NewLocalStorage initializes a new local storage.
func NewLocalStorage() *localStorage {
	return &localStorage{
		Mux:      new(sync.RWMutex),
		Services: make(map[string]*Service),
		Topics:   make(map[string]*Topic),
	}
}

// GetTopic retrieves a topic by its name.
func (s *localStorage) GetTopic(topicName string) (*Topic, error) {
	s.Mux.RLock()
	defer s.Mux.RUnlock()

	topic, exists := s.Topics[topicName]
	if !exists {
		return nil, errors.New("topic not found: " + topicName)
	}

	return topic, nil
}

// GetTopics retrieves all topic names.
func (s *localStorage) GetTopics() []string {
	var topics []string

	s.Mux.RLock()
	defer s.Mux.RUnlock()

	for topicName := range s.Topics {
		topics = append(topics, topicName)
	}

	return topics
}

// AppendTopic adds a new topic to the local storage.
func (s *localStorage) AppendTopic(topicName string, maximumPartitionCount int) {
	topic := &Topic{
		Mux:                   new(sync.Mutex),
		MaximumPartitionCount: maximumPartitionCount,
		Services:              make(map[string]*Service),
	}

	s.Mux.Lock()
	defer s.Mux.Unlock()

	s.Topics[topicName] = topic
	log.Printf("Appended topic: %s with max partitions: %d\n", topicName, maximumPartitionCount)
}

// GetTopicServicePartitions retrieves all partitions for a given topic.
func (s *localStorage) GetTopicServicePartitions(topicName string) ([]string, error) {
	s.Mux.RLock()
	topic, exists := s.Topics[topicName]
	s.Mux.RUnlock()
	if !exists {
		return nil, errors.New("topic not found: " + topicName)
	}

	topic.Mux.Lock()
	defer topic.Mux.Unlock()

	var partitions []string
	for partition := range topic.Services {
		partitions = append(partitions, partition)
	}

	return partitions, nil
}

// AppendService adds a new service to the local storage.
func (s *localStorage) AppendService(service *Service) error {
	topic, err := s.GetTopic(service.Topic)
	if err != nil {
		return err
	}

	topic.Mux.Lock()
	defer topic.Mux.Unlock()

	topic.Services[service.Partition] = service

	s.Mux.Lock()
	defer s.Mux.Unlock()

	s.Services[service.Id] = service
	log.Printf("Appended service: %s to topic: %s at partition: %s\n", service.Id, service.Topic, service.Partition)

	return nil
}

// GetService retrieves a service by its ID or by topic and partition.
func (s *localStorage) GetService(serviceData Service) (*Service, error) {
	if serviceData.Id != "" {
		s.Mux.RLock()
		defer s.Mux.RUnlock()

		service, exists := s.Services[serviceData.Id]
		if !exists {
			return nil, errors.New("service not found: " + serviceData.Id)
		}
		return service, nil
	} else if serviceData.Topic != "" && serviceData.Partition != "" {
		s.Mux.RLock()
		topic, exists := s.Topics[serviceData.Topic]
		s.Mux.RUnlock()
		if !exists {
			return nil, errors.New("topic not found: " + serviceData.Topic)
		}

		topic.Mux.Lock()
		defer topic.Mux.Unlock()

		service, exists := topic.Services[serviceData.Partition]
		if !exists {
			return nil, errors.New("service not found in partition: " + serviceData.Partition)
		}
		return service, nil
	} else {
		return nil, errors.New("invalid service data: either ID or both Topic and Partition must be provided")
	}
}
