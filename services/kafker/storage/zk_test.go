package storage_test

import (
	"testing"
	"time"

	Storage "peergrine/kafker/storage"

	"github.com/go-zookeeper/zk"
	"github.com/stretchr/testify/assert"
)

// 連接到實際的 Zookeeper
func connectToZookeeper(t *testing.T) *zk.Conn {
	servers := []string{"127.0.0.1:2181"}
	conn, _, err := zk.Connect(servers, time.Second*10)
	if err != nil {
		t.Fatalf("Failed to connect to Zookeeper: %v", err)
	}
	return conn
}

// 測試初始化 zkStorage
func TestNewZkStorage(t *testing.T) {
	conn := connectToZookeeper(t)
	defer conn.Close()

	storage, err := Storage.NewZkStorage(conn)
	assert.NoError(t, err)
	assert.NotNil(t, storage)
}

// 測試添加 Topic
func TestAppendTopic(t *testing.T) {
	conn := connectToZookeeper(t)
	defer conn.Close()

	storage, err := Storage.NewZkStorage(conn)
	assert.NoError(t, err)

	// 測試創建 Topic
	topicName := "test-topic"
	err = storage.AppendTopic(topicName, 10)
	assert.NoError(t, err)

	// 確認 Topic 已創建
	topic, err := storage.GetTopic(topicName)
	assert.NoError(t, err)
	assert.NotNil(t, topic)
	assert.Equal(t, 10, topic.MaximumPartitionCount)
}

// 測試獲取 Topic
func TestGetTopic(t *testing.T) {
	conn := connectToZookeeper(t)
	defer conn.Close()

	storage, err := Storage.NewZkStorage(conn)
	assert.NoError(t, err)

	// 假設已經有一個 topic
	topicName := "existing-topic"
	err = storage.AppendTopic(topicName, 5)
	assert.NoError(t, err)

	topic, err := storage.GetTopic(topicName)
	assert.NoError(t, err)
	assert.NotNil(t, topic)
	assert.Equal(t, 5, topic.MaximumPartitionCount)
}

// 測試鎖機制（讀鎖）
func TestRLock(t *testing.T) {
	conn := connectToZookeeper(t)
	defer conn.Close()

	storage, err := Storage.NewZkStorage(conn)
	assert.NoError(t, err)
	lockPath := "/kafker"

	// 獲取讀鎖
	rLock, err := storage.RLock(lockPath)
	assert.NoError(t, err)
	assert.NotNil(t, rLock)

	// 解鎖
	err = rLock.RUnlock()
	assert.NoError(t, err)
}

// 測試鎖機制（寫鎖）
func TestWLock(t *testing.T) {
	conn := connectToZookeeper(t)
	defer conn.Close()

	storage, err := Storage.NewZkStorage(conn)
	assert.NoError(t, err)
	lockPath := "/kafker"

	// 獲取寫鎖
	wLock, err := storage.WLock(lockPath)
	assert.NoError(t, err)
	assert.NotNil(t, wLock)

	// 解鎖
	err = wLock.WUnlock()
	assert.NoError(t, err)
}

// 測試添加 Service Name
func TestAppendServiceName(t *testing.T) {
	conn := connectToZookeeper(t) // 使用你的連接函數
	defer conn.Close()

	storage, err := Storage.NewZkStorage(conn)
	assert.NoError(t, err)

	serviceName := "test-service"

	// 嘗試添加服務名稱
	err = storage.AppendServiceName(serviceName)
	assert.NoError(t, err, "should not return error when appending service name")
}

// 測試添加 Service
func TestAppendService(t *testing.T) {
	conn := connectToZookeeper(t)
	defer conn.Close()

	storage, err := Storage.NewZkStorage(conn)
	assert.NoError(t, err)

	service := &Storage.Service{
		Id:        "service-1",
		Name:      "test-service",
		Topic:     "test-topic",
		Partition: "0",
	}

	// 添加 service
	err = storage.AppendService(service)
	assert.NoError(t, err)

	// 確認 service 已添加
	svc, err := storage.GetService(Storage.Service{Id: "service-1"})
	assert.NoError(t, err)
	assert.NotNil(t, svc)
	assert.Equal(t, "test-service", svc.Name)
}

// 測試確認 Service 存在
func TestGetService(t *testing.T) {
	conn := connectToZookeeper(t)
	defer conn.Close()

	storage, err := Storage.NewZkStorage(conn)
	assert.NoError(t, err)

	// 確保 service 已添加
	svc, err := storage.GetService(Storage.Service{Id: "service-1"})
	assert.NoError(t, err)
	assert.NotNil(t, svc)
	assert.Equal(t, "test-service", svc.Name)
}

// 測試移除 Service
func TestRemoveService(t *testing.T) {
	conn := connectToZookeeper(t)
	defer conn.Close()

	storage, err := Storage.NewZkStorage(conn)
	assert.NoError(t, err)

	// 確保服務存在，然後移除
	err = storage.RemoveService("service-1")
	assert.NoError(t, err)

	// 確認 service 已移除
	svc, err := storage.GetService(Storage.Service{Id: "service-1"})
	assert.Error(t, err)
	assert.Nil(t, svc)
}
