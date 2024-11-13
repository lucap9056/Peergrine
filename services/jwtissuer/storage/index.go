package storage

import (
	"errors"
	Redis "peergrine/utils/redis"
	"sync"
	"time"
)

type Storage struct {
	ServiceId     string
	mux           sync.RWMutex
	refreshTokens map[string]string
	redis         *Redis.Manager
	secret        []byte
}

// New 創建一個新的 Storage 實例。
// 參數:
//
//	ServiceId (string): 管理單元的名稱。
//	redisAddr (string): Redis 伺服器的地址。如果為空字符串，則不使用 Redis。
//
// 返回值:
//
//	*Storage: 初始化的 Storage 實例。
func New(ServiceId, redisAddr string) (*Storage, error) {
	storage := &Storage{
		ServiceId:     ServiceId,
		refreshTokens: make(map[string]string),
	}

	if redisAddr != "" {
		redis, err := Redis.New(redisAddr)
		if err != nil {
			return nil, err
		}
		storage.redis = redis
	}

	return storage, nil
}

// SaveToken 儲存刷新令牌和使用者ID。如果 Redis 可用，則將其儲存到 Redis 中；否則儲存到本地存儲。
// 參數:
//
//	refreshToken (string): 要儲存的刷新令牌。
//	userID (string): 與刷新令牌相關聯的使用者ID。
func (storage *Storage) SaveToken(refreshToken, userID string, duration time.Duration) {
	if storage.redis != nil {
		err := storage.saveTokenInRedis(refreshToken, userID, duration)
		if err == nil {
			return
		}
	}
	storage.saveTokenInLocal(refreshToken, userID)
}

// DeleteToken 刪除指定的刷新令牌。如果 Redis 可用，則從 Redis 中刪除；否則從本地存儲中刪除。
// 參數:
//
//	refreshToken (string): 要刪除的刷新令牌。
func (storage *Storage) DeleteToken(refreshToken string) {
	if storage.redis != nil {
		storage.deleteTokenInRedis(refreshToken)
	}
	storage.deleteTokenInLocal(refreshToken)
}

// GetUserIdFromRefreshToken 從存儲中檢索指定的刷新令牌對應的使用者ID。如果 Redis 可用，則從 Redis 中檢索；否則從本地存儲中檢索。
// 參數:
//
//	refreshToken (string): 要檢索的刷新令牌。
//
// 返回值:
//
//	string: 與給定刷新令牌相關聯的使用者ID。如果令牌不存在，則返回空字符串。
//	bool: 如果找到對應的使用者ID，則返回 true，否則返回 false。
func (storage *Storage) GetUserIdFromRefreshToken(refreshToken string) (string, bool) {
	if storage.redis != nil {
		return storage.getRefreshTokenFromRedis(refreshToken)
	}
	return storage.getRefreshTokenFromLocal(refreshToken)
}

// SaveSecret 儲存秘密字串。如果 Redis 可用，則將其儲存到 Redis 中；否則儲存到本地存儲。
// 參數:
//
//	secret ([]byte): 要儲存的秘密字串。
//
// 返回值:
//
//	error: 儲存過程中的錯誤，如果沒有錯誤，則返回 nil。
func (storage *Storage) SaveSecret(secret []byte) error {
	storage.secret = secret
	if storage.redis != nil {
		return storage.saveSecretInRedis(secret)
	}
	return nil
}

// GetSecret 從存儲中檢索指定單位名稱的秘密字串。如果 Redis 可用，則從 Redis 中檢索；否則返回空字節和錯誤。
// 參數:
//
//	ServiceId (string): 要檢索秘密字串的單位名稱。
//
// 返回值:
//
//	[]byte: 檢索到的秘密字串。如果秘密字串不存在，則返回 nil。
//	error: 檢索過程中的錯誤，如果沒有錯誤，則返回 nil。
func (storage *Storage) GetSecret(ServiceId string) ([]byte, error) {

	if ServiceId == storage.ServiceId {
		return storage.secret, nil
	}

	if storage.redis != nil {
		return storage.getSecretInRedis(ServiceId)
	}

	return nil, errors.New("secret not found in local storage")
}

// Close 關閉 Redis 連接，並刪除指定單位名稱的秘密字串。如果不使用 Redis，則返回 nil。
// 返回值:
//
//	error: 關閉過程中的錯誤，如果沒有錯誤，則返回 nil。
func (storage *Storage) Close() error {
	if storage.redis != nil {
		return storage.deleteSecretInRedis(storage.ServiceId)
	}
	return nil
}
