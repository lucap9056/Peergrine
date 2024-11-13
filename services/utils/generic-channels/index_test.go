package genericchannels_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	GenericChannels "peergrine/utils/generic-channels"

	"github.com/stretchr/testify/assert"
)

// 測試添加新信號通道
func TestAddSignalChannel(t *testing.T) {
	sc := GenericChannels.New[int]()
	key := "testKey"

	// 添加信號通道
	ch := sc.Add(key)
	assert.NotNil(t, ch, "Channel should be created and returned")

	// 確認通道是否已正確存儲
	retrievedCh := sc.Get(key)
	assert.Equal(t, ch, retrievedCh, "Retrieved channel should match the created one")
}

// 測試獲取不存在的信號通道
func TestGetNonExistentChannel(t *testing.T) {
	sc := GenericChannels.New[int]()

	// 嘗試獲取不存在的通道
	ch := sc.Get("nonExistentKey")
	assert.Nil(t, ch, "Should return nil for a non-existent channel")
}

// 測試刪除信號通道
func TestDelSignalChannel(t *testing.T) {
	sc := GenericChannels.New[int]()
	key := "testKey"

	// 添加並刪除信號通道
	sc.Add(key)
	sc.Del(key)

	// 確認通道是否已刪除
	ch := sc.Get(key)
	assert.Nil(t, ch, "Channel should be nil after deletion")
}

// 測試刪除後通道的行為
func TestDelSignalChannelClosesChannel(t *testing.T) {
	sc := GenericChannels.New[int]()
	key := "testKey"

	// 添加信號通道
	ch := sc.Add(key)

	// 確認通道被關閉後的行為
	sc.Del(key)

	// 向關閉的通道發送數據應該會引發 panic
	assert.Panics(t, func() {
		ch <- 0
	}, "Should panic when sending to a closed channel")
}

// 測試並發訪問通道
func TestConcurrentAccess(t *testing.T) {
	sc := GenericChannels.New[int]()
	var wg sync.WaitGroup

	// 並發創建和刪除通道
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("testKey%d", i)
			sc.Add(key)
			time.Sleep(time.Millisecond * 10) // 模擬處理時間
			sc.Del(key)
		}(i)
	}

	wg.Wait()

	// 最後檢查是否所有通道已刪除
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("testKey%d", i)
		ch := sc.Get(key)
		assert.Nil(t, ch, "Channel should be nil after concurrent deletion")
	}
}
