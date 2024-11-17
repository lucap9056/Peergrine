package redis_test

import (
	"testing"
	"time"

	"peergrine/utils/redis"

	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

func TestManager_Get(t *testing.T) {
	client, mock := redismock.NewClientMock()

	manager := redis.Test(client)
	mock.ExpectGet("key1").SetVal("value1")

	val, err := manager.Get("key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", string(val))

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestManager_Set(t *testing.T) {
	client, mock := redismock.NewClientMock()

	manager := redis.Test(client)
	value := []byte("value1")
	mock.ExpectSet("key1", value, time.Minute).SetVal("OK")

	err := manager.Set("key1", value, time.Minute)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestManager_Exists(t *testing.T) {
	client, mock := redismock.NewClientMock()

	manager := redis.Test(client)
	mock.ExpectExists("key1").SetVal(1)

	exists, err := manager.Exists("key1")
	assert.NoError(t, err)
	assert.True(t, exists)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestManager_Del(t *testing.T) {
	client, mock := redismock.NewClientMock()

	manager := redis.Test(client)
	mock.ExpectDel("key1").SetVal(1)

	err := manager.Del("key1")
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestManager_Scan(t *testing.T) {
	client, mock := redismock.NewClientMock()

	manager := redis.Test(client)
	mock.ExpectScan(uint64(0), "*", int64(10)).SetVal([]string{"key1", "key2"}, uint64(0))

	keys, cursor, err := manager.Scan(0, "*", 10)
	assert.NoError(t, err)
	assert.Equal(t, []string{"key1", "key2"}, keys)
	assert.Equal(t, uint64(0), cursor)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestManager_MGet(t *testing.T) {
	client, mock := redismock.NewClientMock()

	manager := redis.Test(client)
	mock.ExpectMGet("key1", "key2").SetVal([]interface{}{"value1", "value2"})

	values, err := manager.MGet([]string{"key1", "key2"})
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{"value1", "value2"}, values)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestManager_Close(t *testing.T) {
	client, mock := redismock.NewClientMock()

	manager := redis.Test(client)
	mock.ClearExpect()

	err := manager.Close()
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
