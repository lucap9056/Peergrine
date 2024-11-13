package redis

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	addr := "127.0.0.1:6379"
	manager, err := New(addr)
	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, addr, manager.client.Options().Addr)
}

func TestGet(t *testing.T) {
	db, mock := redismock.NewClientMock()
	manager := Test(context.Background(), db)

	key := "testkey"
	expectedValue := []byte("testvalue")

	mock.ExpectGet(key).SetVal(string(expectedValue))

	value, err := manager.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, value)
}

func TestSet(t *testing.T) {
	db, mock := redismock.NewClientMock()
	manager := Test(context.Background(), db)

	key := "testkey"
	value := []byte("testvalue")
	expiration := time.Minute

	mock.ExpectSet(key, value, expiration).SetVal("OK")

	err := manager.Set(key, value, expiration)
	assert.NoError(t, err)
}

func TestExist(t *testing.T) {
	db, mock := redismock.NewClientMock()
	manager := Test(context.Background(), db)

	key := "testkey"

	mock.ExpectExists(key).SetVal(1)

	exists, err := manager.Exists(key)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestDel(t *testing.T) {
	db, mock := redismock.NewClientMock()
	manager := Test(context.Background(), db)

	key := "testkey"

	mock.ExpectDel(key).SetVal(1)

	err := manager.Del(key)
	assert.NoError(t, err)
}

func TestClose(t *testing.T) {
	db, _ := redismock.NewClientMock()
	manager := Test(context.Background(), db)

	err := manager.Close()
	assert.NoError(t, err)
}
