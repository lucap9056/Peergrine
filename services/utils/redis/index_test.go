package redis_test

import (
	"testing"

	"peergrine/utils/redis"

	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_RetryOnPingFailure(t *testing.T) {

	client, mock := redismock.NewClientMock()

	mock.ExpectPing().RedisNil()
	mock.ExpectPing().RedisNil()
	mock.ExpectPing().SetVal("PONG")

	manager, err := redis.Test(client)

	require.NoError(t, err)
	assert.NotNil(t, manager)
}

func TestManager_SetGet(t *testing.T) {

	client, mock := redismock.NewClientMock()

	mock.ExpectPing().SetVal("PONG")

	mock.ExpectClusterInfo().RedisNil()

	manager, err := redis.Test(client)
	require.NoError(t, err)

	mock.ExpectSet("test_key", []byte("test_value"), 0).SetVal("OK")
	err = manager.Set("test_key", []byte("test_value"), 0)
	assert.NoError(t, err)

	mock.ExpectGet("test_key").SetVal("test_value")
	val, err := manager.Get("test_key")
	assert.NoError(t, err)
	assert.Equal(t, []byte("test_value"), val)
}

func TestManager_Del(t *testing.T) {

	client, mock := redismock.NewClientMock()

	mock.ExpectPing().SetVal("PONG")

	mock.ExpectClusterInfo().RedisNil()

	manager, err := redis.Test(client)
	require.NoError(t, err)

	mock.ExpectDel("test_key").SetVal(1)
	err = manager.Del("test_key")
	assert.NoError(t, err)
}

func TestManager_Exists(t *testing.T) {

	client, mock := redismock.NewClientMock()

	mock.ExpectPing().SetVal("PONG")

	mock.ExpectClusterInfo().RedisNil()

	manager, err := redis.Test(client)
	require.NoError(t, err)

	mock.ExpectExists("test_key").SetVal(1)
	exists, err := manager.Exists("test_key")
	assert.NoError(t, err)
	assert.True(t, exists)
}
