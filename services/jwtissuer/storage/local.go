package storage

// saveTokenInLocal 將刷新令牌和使用者ID儲存到本地存儲中。
// 參數:
//
//	refreshToken (string): 要儲存的刷新令牌。
//	userID (string): 與刷新令牌相關聯的使用者ID。
func (storage *Storage) saveTokenInLocal(refreshToken, userID string) {
	storage.mux.Lock()
	defer storage.mux.Unlock()
	storage.refreshTokens[refreshToken] = userID
}

// deleteTokenInLocal 從本地存儲中刪除指定的刷新令牌。
// 參數:
//
//	refreshToken (string): 要刪除的刷新令牌。
func (storage *Storage) deleteTokenInLocal(refreshToken string) {
	storage.mux.Lock()
	defer storage.mux.Unlock()

	delete(storage.refreshTokens, refreshToken)
}

// getRefreshTokenFromLocal 從本地存儲中檢索指定的刷新令牌對應的使用者ID。
// 參數:
//
//	refreshToken (string): 要檢索的刷新令牌。
//
// 返回值:
//
//	string: 與給定刷新令牌相關聯的使用者ID。如果令牌不存在，則返回空字符串。
//	bool: 如果找到對應的使用者ID，則返回 true，否則返回 false。
func (storage *Storage) getRefreshTokenFromLocal(refreshToken string) (string, bool) {
	t, b := storage.refreshTokens[refreshToken]
	return t, b
}
