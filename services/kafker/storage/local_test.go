package storage_test

import (
	"testing"

	Storage "peergrine/kafker/storage"

	"github.com/stretchr/testify/assert"
)

func TestLocalStorage(t *testing.T) {
	storage := Storage.NewLocalStorage()

	// 測試添加主題
	t.Run("AppendTopic", func(t *testing.T) {
		topicName := "test-topic"
		maxPartitionCount := 5

		storage.AppendTopic(topicName, maxPartitionCount)

		topic, err := storage.GetTopic(topicName)
		assert.NoError(t, err)
		assert.NotNil(t, topic)
		assert.Equal(t, maxPartitionCount, topic.MaximumPartitionCount)
	})

	// 測試獲取所有主題
	t.Run("GetTopics", func(t *testing.T) {
		topics := storage.GetTopics()
		assert.Contains(t, topics, "test-topic")
	})

	// 測試添加服務
	t.Run("AppendService", func(t *testing.T) {
		service := &Storage.Service{
			Id:        "service-1",
			Name:      "test-service",
			Topic:     "test-topic",
			Partition: "0",
		}

		err := storage.AppendService(service)
		assert.NoError(t, err)

		topic, err := storage.GetTopic("test-topic")
		assert.NoError(t, err)
		assert.NotNil(t, topic)
		assert.Contains(t, topic.Services, "0")
	})

	// 測試獲取服務
	t.Run("GetServiceById", func(t *testing.T) {
		service, err := storage.GetService(Storage.Service{Id: "service-1"})
		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, "test-service", service.Name)
	})

	t.Run("GetServiceByTopicAndPartition", func(t *testing.T) {
		service, err := storage.GetService(Storage.Service{Topic: "test-topic", Partition: "0"})
		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, "test-service", service.Name)
	})

	// 測試獲取不存在的主題
	t.Run("GetNonExistentTopic", func(t *testing.T) {
		_, err := storage.GetTopic("non-existent-topic")
		assert.Error(t, err)
	})

	// 測試獲取不存在的服務
	t.Run("GetNonExistentService", func(t *testing.T) {
		_, err := storage.GetService(Storage.Service{Id: "non-existent-service"})
		assert.Error(t, err)
	})
}
