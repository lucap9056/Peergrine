package auth

type TokenData struct {
	Token  string
	Iss    string `json:"iss"`
	Iat    int64  `json:"iat"`
	Exp    int64  `json:"exp"`
	UserId string `json:"user_id"`
}

func (t *TokenData) SetToken(token string) {
	t.Token = token
}
