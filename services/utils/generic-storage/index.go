package genericstorage

import (
	Auth "peergrine/utils/auth"
	Redis "peergrine/utils/redis"
	"sync"
)

// base interface defines the minimum methods that data types managed by Storage must implement.
type base interface {
	GetKey() string
	GetExpiresAt() int64
}

// Storage manages both local in-memory storage and Redis-based storage.
// It handles caching of tokens and secrets in a thread-safe manner.
type Storage[T base] struct {
	ChannelId int32
	Redis     *Redis.Manager
	Local     *LocalStorageManager[T]
	SecretMux *sync.RWMutex
	Secrets   map[string][]byte
}

// New creates and returns a new instance of Storage.
// It initializes the Redis manager if a Redis address is provided.
func New[T base](channelId int32, redisAddr string) (*Storage[T], error) {
	manager := &Storage[T]{
		ChannelId: channelId,
		SecretMux: new(sync.RWMutex),
		Secrets:   make(map[string][]byte),
	}

	if redisAddr != "" {
		redis, err := Redis.New(redisAddr)
		if err != nil {
			return nil, err
		}
		manager.Redis = redis
	}

	manager.Local = NewLocalStorageManager[T]()

	return manager, nil
}

// SetTokenCache sets the token data into the local cache.
// Parameters:
//   - token (string): The token string.
//   - tokenData (Auth.TokenData): The token data to be cached.
func (m *Storage[any]) SetTokenCache(token string, tokenData Auth.TokenPayload) {
	tokenData.SetToken(token)
	m.Local.SetToken(tokenData)
}

// GetTokenCache retrieves the token data from the local cache based on the token string.
// Parameters:
//   - token (string): The token string.
//
// Returns:
//   - *Auth.TokenData: The token data if found, or nil if not present.
func (m *Storage[any]) GetTokenCache(token string) *Auth.TokenPayload {
	return m.Local.GetToken(token)
}

// GetSecret retrieves the secret associated with a unit name from either the local cache or Redis.
// If the secret is not found locally, it fetches from Redis and updates the local cache.
// Parameters:
//   - unitName (string): The name of the unit.
//
// Returns:
//   - []byte: The retrieved secret data.
//   - error: An error if the operation fails, otherwise nil.
func (m *Storage[any]) GetSecret(unitName string) ([]byte, error) {

	{
		m.SecretMux.RLock()
		secret, exists := m.Secrets[unitName]
		m.SecretMux.RUnlock()
		if exists {
			return secret, nil
		}
	}

	{
		key := "secret:" + unitName
		secret, err := m.Redis.Get(key)
		if err != nil {
			return nil, err
		}

		m.SecretMux.Lock()
		m.Secrets[unitName] = secret
		m.SecretMux.Unlock()
		return secret, nil
	}
}

// Close closes the Redis client connection and releases local storage resources.
// Returns:
//   - error: If successful, returns nil, otherwise an error message.
func (m *Storage[any]) Close() error {

	m.Local.Close()

	if m.Redis != nil {
		return m.Redis.Close()
	}
	return nil
}
