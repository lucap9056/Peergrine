package rtcbridgeapi

// Candidate represents a WebRTC candidate.
type Candidate struct {
	Candidate     *string `json:"candidate"`
	SdpMLineIndex *int    `json:"sdpMLineIndex"`
	SdpMid        *string `json:"sdpMid"`
}

// SignalData contains the data of a signal.
type SignalData struct {
	ClientId   string      `json:"client_id"`
	ClientName string      `json:"client_name"`
	ChannelId  string      `json:"channel_id"`
	SDP        string      `json:"sdp"`
	Candidates []Candidate `json:"candidates"`
}

// LinkCode contains the link code and expiration time.
type LinkCode struct {
	LinkCode  string `json:"link_code"`
	ExpiresAt int64  `json:"expires_at"`
}

type KafkerSignal struct {
	LinkCode string     `json:"link_code"`
	Signal   SignalData `json:"signal"`
}
