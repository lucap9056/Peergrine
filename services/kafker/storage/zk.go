package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	ZkRWLock "peergrine/kafker/storage/zk-rwlock"

	"github.com/go-zookeeper/zk"
)

const (
	zkBasePath       = "/kafker"
	zkServiceIdsPath = "/serviceIds"
	zkServicesPath   = "/services"
	zkTopicsPath     = "/topics"
)

type zkStorage struct {
	conn *zk.Conn
}

func NewZkStorage(conn *zk.Conn) (*zkStorage, error) {
	zkPaths := []string{
		zkBasePath,
		zkBasePath + zkServiceIdsPath,
		zkBasePath + zkServicesPath,
		zkBasePath + zkTopicsPath,
	}
	permAll := zk.WorldACL(zk.PermAll)

	for _, path := range zkPaths {
		if _, err := conn.Create(path, nil, 0, permAll); err != nil && err != zk.ErrNodeExists {
			return nil, fmt.Errorf("failed to create zookeeper path %s: %w", path, err)
		}
	}

	log.Println("Zookeeper storage initialized successfully.")
	return &zkStorage{conn: conn}, nil
}

func (s *zkStorage) WLock(nodePath string) (*ZkRWLock.ZkWLock, error) {
	return ZkRWLock.WLock(s.conn, nodePath)
}

func (s *zkStorage) RLock(nodePath string) (*ZkRWLock.ZkRLock, error) {
	return ZkRWLock.RLock(s.conn, nodePath)
}

func (s *zkStorage) AppendTopic(topicName string, maxPartitionCount int) error {
	topicPath := zkBasePath + zkTopicsPath + "/" + topicName
	topic := &Topic{MaximumPartitionCount: maxPartitionCount}

	topicData, err := json.Marshal(topic)
	if err != nil {
		return fmt.Errorf("failed to marshal topic data for %s: %w", topicName, err)
	}

	if _, err := s.conn.Create(topicPath, topicData, 0, zk.WorldACL(zk.PermAll)); err != nil {
		return fmt.Errorf("failed to create topic %s: %w", topicName, err)
	}

	log.Printf("Successfully added topic: %s", topicName)
	return nil
}

func (s *zkStorage) GetTopic(topicName string) (*Topic, error) {
	topicPath := zkBasePath + zkTopicsPath + "/" + topicName
	topicData, _, err := s.conn.Get(topicPath)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve topic %s: %w", topicName, err)
	}

	var topic Topic
	if err := json.Unmarshal(topicData, &topic); err != nil {
		return nil, fmt.Errorf("failed to unmarshal topic data for %s: %w", topicName, err)
	}

	return &topic, nil
}

func (s *zkStorage) GetTopics() ([]string, error) {
	children, _, err := s.conn.Children(zkBasePath + zkTopicsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get children of topics: %w", err)
	}

	var topicNames []string
	for _, node := range children {
		if !strings.Contains(node, ZkRWLock.ZKRLOCK) && !strings.Contains(node, ZkRWLock.ZKWLOCK) {
			topicNames = append(topicNames, node)
		}
	}

	return topicNames, nil
}

func (s *zkStorage) GetTopicServicePartitions(topicName string) ([]string, error) {
	topicPath := zkBasePath + zkTopicsPath + "/" + topicName
	partitions, _, err := s.conn.Children(topicPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get partitions for topic %s: %w", topicName, err)
	}

	var partitionNames []string
	for _, node := range partitions {
		if !strings.Contains(node, ZkRWLock.ZKRLOCK) && !strings.Contains(node, ZkRWLock.ZKWLOCK) {
			partitionNames = append(partitionNames, node)
		}
	}

	return partitionNames, nil
}

func (s *zkStorage) AppendServiceName(serviceName string) error {
	servicePath := zkBasePath + zkServicesPath + "/" + serviceName
	if _, err := s.conn.Create(servicePath, nil, 0, zk.WorldACL(zk.PermAll)); err != nil && err != zk.ErrNodeExists {
		return fmt.Errorf("failed to append service name %s: %w", serviceName, err)
	}

	log.Printf("Successfully added service name: %s", serviceName)
	return nil
}

func (s *zkStorage) AppendService(service *Service) error {
	serviceData, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("failed to marshal service data for %v: %w", service, err)
	}

	permAll := zk.WorldACL(zk.PermAll)
	serviceNamePath := zkBasePath + zkServicesPath + "/" + service.Name

	if _, err := s.conn.Create(serviceNamePath, nil, 0, permAll); err != nil && err != zk.ErrNodeExists {
		return fmt.Errorf("failed to create service name path %s: %w", serviceNamePath, err)
	}

	serviceTopicPath := zkBasePath + zkTopicsPath + "/" + service.Topic + "/" + service.Partition
	serviceNameIDPath := zkBasePath + zkServicesPath + "/" + service.Name + "/" + service.Id
	serviceIdPath := zkBasePath + zkServiceIdsPath + "/" + service.Id

	log.Printf("Preparing to write service data for service ID %s", service.Id)

	exists, _, err := s.conn.Exists(serviceIdPath)
	if err != nil {
		return fmt.Errorf("failed to check service exists path %s: %w", serviceNamePath, err)
	}

	if exists {
		if err := s.RemoveService(service.Id); err != nil {
			return err
		}
	}

	reqs := []interface{}{
		&zk.CreateRequest{Path: serviceTopicPath, Data: serviceData, Acl: permAll, Flags: 0},
		&zk.CreateRequest{Path: serviceNameIDPath, Data: serviceData, Acl: permAll, Flags: 0},
		&zk.CreateRequest{Path: serviceIdPath, Data: serviceData, Acl: permAll, Flags: 0},
	}

	if _, err := s.conn.Multi(reqs...); err != nil {
		return fmt.Errorf("failed to append service data: %w", err)
	}

	log.Printf("Successfully added service: %s", service.Id)
	return nil
}

func (s *zkStorage) GetService(serviceData Service) (*Service, error) {
	var path string

	if serviceData.Id != "" {
		path = zkBasePath + zkServiceIdsPath + "/" + serviceData.Id
	} else if serviceData.Topic != "" && serviceData.Partition != "" {
		path = zkBasePath + zkTopicsPath + "/" + serviceData.Topic + "/" + serviceData.Partition
	} else {
		return nil, errors.New("neither ID nor topic/partition specified")
	}

	serviceDataBytes, _, err := s.conn.Get(path)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil, fmt.Errorf("service not found at path %s: %w", path, err)
		}
		return nil, fmt.Errorf("failed to retrieve service data from path %s: %w", path, err)
	}

	var service Service
	if err := json.Unmarshal(serviceDataBytes, &service); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service data: %w", err)
	}

	return &service, nil
}

func (s *zkStorage) GetAllServices() ([]string, error) {
	serviceIdsPath := zkBasePath + zkServiceIdsPath
	children, _, err := s.conn.Children(serviceIdsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get all service IDs: %w", err)
	}

	return children, nil
}

func (s *zkStorage) GetServiceNames() ([]string, error) {
	serviceNamesPath := zkBasePath + zkServicesPath
	children, _, err := s.conn.Children(serviceNamesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get service names: %w", err)
	}

	return children, nil
}

func (s *zkStorage) RemoveService(serviceId string) error {
	service, err := s.GetService(Service{Id: serviceId})
	if err != nil {
		return fmt.Errorf("failed to get service with ID %s: %w", serviceId, err)
	}

	serviceTopicPath := zkBasePath + zkTopicsPath + "/" + service.Topic + "/" + service.Partition
	serviceNamePath := zkBasePath + zkServicesPath + "/" + service.Name + "/" + service.Id
	serviceIdPath := zkBasePath + zkServiceIdsPath + "/" + service.Id

	log.Printf("Preparing to delete service data for service ID %s", serviceId)

	_, stat, err := s.conn.Exists(serviceIdPath)
	if err != nil {
		return fmt.Errorf("failed to check existence of service node %s: %w", serviceIdPath, err)
	}

	if stat != nil {
		creationTime := time.Unix(0, stat.Ctime*int64(time.Millisecond))
		if time.Since(creationTime) < time.Minute {
			log.Printf("Service %s was created less than 60 seconds ago, skipping deletion.", serviceId)
			return nil
		}
	}

	reqs := []interface{}{
		&zk.DeleteRequest{Path: serviceTopicPath, Version: -1},
		&zk.DeleteRequest{Path: serviceNamePath, Version: -1},
		&zk.DeleteRequest{Path: serviceIdPath, Version: -1},
	}

	if _, err := s.conn.Multi(reqs...); err != nil {
		return fmt.Errorf("failed to remove service %s: %w", serviceId, err)
	}

	if err := s.removeEmptyNodes(service.Topic, zkBasePath+zkTopicsPath); err != nil {
		return err
	}
	if err := s.removeEmptyNodes(service.Name, zkBasePath+zkServicesPath); err != nil {
		return err
	}

	log.Printf("Successfully removed service: %s", serviceId)
	return nil
}

func (s *zkStorage) removeEmptyNodes(nodeName, basePath string) error {
	children, _, err := s.conn.Children(basePath + "/" + nodeName)
	if err != nil {
		return fmt.Errorf("failed to check children of %s: %w", nodeName, err)
	}

	if len(children) == 0 {
		if err := s.conn.Delete(basePath+"/"+nodeName, -1); err != nil {
			return fmt.Errorf("failed to delete empty node %s: %w", nodeName, err)
		}
		log.Printf("Deleted empty node: %s", nodeName)
	}
	return nil
}
