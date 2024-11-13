package storage

import (
	Keys "peergrine/jwtissuer/storage/keys"
	"time"
)

// saveTokenInRedis 將刷新令牌和使用者ID儲存到 Redis 中。
// 參數:
//
//	refreshToken (string): 要儲存的刷新令牌。
//	userID (string): 與刷新令牌相關聯的使用者ID。
//
// 返回值:
//
//	error: 儲存過程中的錯誤，如果沒有錯誤，則返回 nil。
func (storage *Storage) saveTokenInRedis(refreshToken, userId string, duration time.Duration) error {
	key := Keys.RefreshToken(refreshToken)
	userIdBytes := []byte(userId)

	return storage.redis.Set(key, userIdBytes, duration)
}

// deleteTokenInRedis 從 Redis 中刪除指定的刷新令牌。
// 參數:
//
//	refreshToken (string): 要刪除的刷新令牌。
//
// 返回值:
//
//	error: 刪除過程中的錯誤，如果沒有錯誤，則返回 nil。
func (storage *Storage) deleteTokenInRedis(refreshToken string) error {
	key := Keys.RefreshToken(refreshToken)

	return storage.redis.Del(key)
}

// getRefreshTokenFromRedis 從 Redis 中檢索指定的刷新令牌對應的使用者ID。
// 參數:
//
//	refreshToken (string): 要檢索的刷新令牌。
//
// 返回值:
//
//	string: 與給定刷新令牌相關聯的使用者ID。如果令牌不存在，則返回空字符串。
//	bool: 如果找到對應的使用者ID，則返回 true，否則返回 false。
func (storage *Storage) getRefreshTokenFromRedis(refreshToken string) (string, bool) {
	key := Keys.RefreshToken(refreshToken)
	userIdBytes, err := storage.redis.Get(key)
	if err != nil {
		return "", false
	}

	return string(userIdBytes), true
}

// saveSecretInRedis 將秘密字串儲存到 Redis 中，並將其編碼為 base64。
// 參數:
//
//	secret ([]byte): 要儲存的秘密字串。
//
// 返回值:
//
//	error: 儲存過程中的錯誤，如果沒有錯誤，則返回 nil。
func (storage *Storage) saveSecretInRedis(secret []byte) error {
	key := Keys.Secret(storage.ServiceId)
	return storage.redis.Set(key, secret, 0)
}

// getSecretInRedis 從 Redis 中檢索指定單位名稱的秘密字串，並將其解碼為原始字節。
// 參數:
//
//	ServiceId (string): 要檢索秘密字串的單位名稱。
//
// 返回值:
//
//	[]byte: 解碼後的秘密字串。如果秘密字串不存在，則返回 nil。
//	error: 檢索過程中的錯誤，如果沒有錯誤，則返回 nil。
func (storage *Storage) getSecretInRedis(ServiceId string) ([]byte, error) {
	key := Keys.Secret(ServiceId)
	secret, err := storage.redis.Get(key)
	if err == nil {
		return nil, err
	}

	return secret, nil
}

// deleteSecretInRedis 從 Redis 中刪除指定單位名稱的秘密字串。
// 參數:
//
//	ServiceId (string): 要刪除秘密字串的單位名稱。
//
// 返回值:
//
//	error: 刪除過程中的錯誤，如果沒有錯誤，則返回 nil。
func (storage *Storage) deleteSecretInRedis(ServiceId string) error {
	key := Keys.Secret(ServiceId)
	return storage.redis.Del(key)
}
