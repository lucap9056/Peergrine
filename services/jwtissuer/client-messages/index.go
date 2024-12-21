package clientmessages

type Message[T any] struct {
	Type    string `json:"type"`
	Content T      `json:"content"`
}

type AuthorizationMessage struct {
	RefreshToken string `json:"refresh_token"`
	BearerToken  string `json:"access_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// Authorization 創建一個授權消息的 Message 實例。
// 參數:
//
//	refreshToken (string): 刷新令牌。
//	bearerToken (string): 存取令牌。
//	expiresAt (int64): 存取令牌的過期時間（以時間戳表示）。
//
// 返回值:
//
//	Message[AuthorizationMessage]: 包含授權資訊的 Message 實例。
func Authorization(refreshToken, bearerToken string, expiresAt int64) Message[AuthorizationMessage] {
	return Message[AuthorizationMessage]{
		Type: "Authorization",
		Content: AuthorizationMessage{
			RefreshToken: refreshToken,
			BearerToken:  bearerToken,
			ExpiresAt:    expiresAt,
		},
	}
}
