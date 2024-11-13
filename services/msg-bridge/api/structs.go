package msgbridgeapi

// LinkCode contains the link code and expiration time.
type LinkCode struct {
	LinkCode  string `json:"link_code"`
	ExpiresAt int64  `json:"expires_at"`
}

type SessionData struct {
	ClientId  string `json:"client_id"`
	PublicKey string `json:"public_key"`
}

type MessageData struct {
	SenderId string `json:"sender_id"`
	Message  string `json:"message"`
}

type KafkaMessage struct {
	ClientId string `json:"client_id"`
	Message  []byte `json:"message"`
}
