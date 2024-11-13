package genericstorage

import (
	"encoding/json"
	"time"
)

// SetToRedis stores data of type T into Redis with an expiration time.
// Parameters:
//   - key (string): The key under which the data will be stored in Redis.
//   - data (T): The data object that implements the base interface.
//
// Returns:
//   - error: If the operation is successful, returns nil, otherwise returns an error.
func (m *Storage[T]) SetToRedis(key string, data T) error {

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	duration := time.Duration(data.GetExpiresAt()-time.Now().Unix()) * time.Second
	return m.Redis.Set(key, dataBytes, duration)
}

// GetFromRedis retrieves data of type T from Redis based on the provided key.
// Parameters:
//   - key (string): The key under which the data is stored in Redis.
//
// Returns:
//   - *T: The data retrieved from Redis if found, or nil if an error occurs.
//   - error: If the operation is successful, returns nil, otherwise returns an error.
func (m *Storage[T]) GetFromRedis(key string) (*T, error) {

	result, err := m.Redis.Get(key)
	if err != nil {
		return nil, err
	}

	var data T
	if err := json.Unmarshal(result, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
