package auth

type TokenPayload struct {
	Token     string
	Iss       string `json:"iss"`
	Iat       int64  `json:"iat"`
	Exp       int64  `json:"exp"`
	UserId    string `json:"user_id"`
	ChannelId string `json:"channel_id"`
}

func (t *TokenPayload) SetToken(token string) {
	t.Token = token
}
