package rtcbridgeapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	ServiceAuth "peergrine/grpc/serviceauth"
	AuthMessage "peergrine/jwtissuer/client-messages"
	Storage "peergrine/rtc-bridge/storage"
	Auth "peergrine/utils/auth"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Constants
const LINK_CODE_LETTER_BYTES = "abcdefghijkmnopqrstuvwxyzABCDEFHJKLMNPQRSTUVWXYZ0123456789"
const LINK_CODE_DURATION = 5 * time.Minute

const MESSAGE_TYPE = "signaling"

func (app *API) generateUniqueLinkCode() (string, error) {
	const maxAttempts = 5
	const linkCodeLength = 8

	letterLen := len(LINK_CODE_LETTER_BYTES)

	for attempts := 0; attempts < maxAttempts; attempts++ {
		b := make([]byte, linkCodeLength)
		for i := range b {
			randNum, err := rand.Int(rand.Reader, big.NewInt(int64(letterLen)))
			if err != nil {
				return "", err
			}
			b[i] = LINK_CODE_LETTER_BYTES[randNum.Int64()]
		}

		linkCode := string(b)

		exist, err := app.storage.SignalExists(linkCode)
		if err != nil {
			return "", err
		}
		if !exist {
			return linkCode, nil
		}
	}

	return "", errors.New("failed to generate a unique link code after multiple attempts")
}

func getPlayload(c *gin.Context) (*Auth.TokenPayload, error) {

	tokenPayloadValue, exists := c.Get(TOKEN_PARLOAD)
	if !exists {
		return nil, errors.New("client ID not found in the request context")
	}
	tokenPayload, ok := tokenPayloadValue.(Auth.TokenPayload)
	if !ok {
		return nil, errors.New("invalid format for Client ID")
	}

	return &tokenPayload, nil
}

func (app *API) getChannelId(payload *Auth.TokenPayload) (bool, string) {
	if app.config.UnifiedMessage {
		return true, payload.ChannelId
	}
	return false, app.config.Id
}

func (app *API) setSignal(c *gin.Context) {
	tokenPayload, err := getPlayload(c)
	if err != nil {
		Error(c, http.StatusForbidden, err.Error())
		return
	}

	clientId := tokenPayload.UserId

	var signal SignalData

	if err := c.ShouldBindJSON(&signal); err != nil {
		Error(c, http.StatusBadRequest, fmt.Sprintf("Invalid JSON format: %s", err.Error()))
		return
	}

	if !validSDP(signal.SDP) || !validCandidates(signal.Candidates) {
		Error(c, http.StatusBadRequest, "Invalid SDP or Candidates")
		return
	}

	linkCode, err := app.generateUniqueLinkCode()
	if err != nil {
		Error(c, http.StatusInternalServerError, fmt.Sprintf("failed to generate link code: %v", err))
		return
	}

	signal.ClientId = clientId

	unifiedMessage, channelId := app.getChannelId(tokenPayload)

	signal.ChannelId = channelId

	signalBytes, err := json.Marshal(signal)
	if err != nil {
		Error(c, http.StatusInternalServerError, fmt.Sprintf("failed to marshal signal: %v", err))
		return
	}

	expiresAt := time.Now().Add(LINK_CODE_DURATION).Unix()
	clientSignal := Storage.NewSignal(clientId, linkCode, signalBytes, expiresAt)

	clientSignal.SetChannelId(signal.ChannelId)

	if err := app.storage.SetSignal(clientSignal); err != nil {
		Error(c, http.StatusInternalServerError, fmt.Sprintf("failed to store client signal: %v", err))
		return
	}
	defer app.storage.RemoveSignal(linkCode)

	result := LinkCode{
		LinkCode:  linkCode,
		ExpiresAt: expiresAt,
	}

	resultBytes, _ := json.Marshal(result)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Status(http.StatusOK)
	c.Writer.Write(resultBytes)
	c.Writer.Flush()

	ctx, cancel := context.WithTimeout(context.Background(), LINK_CODE_DURATION)
	defer cancel()

	closeNotify := c.Writer.CloseNotify()

	if unifiedMessage {

		select {
		case <-ctx.Done():
			Error(c, http.StatusRequestTimeout, "Request timed out")
		case <-closeNotify:
		}
	} else {

		awaitSignal := app.signalChannels.Add(linkCode)
		defer app.signalChannels.Del(linkCode)

		select {
		case targetSignal, ok := <-awaitSignal:
			if ok {
				targetSignalBytes, _ := json.Marshal(targetSignal)
				c.Writer.Write(targetSignalBytes)
				c.Writer.Flush()
			}
		case <-ctx.Done():
			Error(c, http.StatusRequestTimeout, "Request timed out")
		case <-closeNotify:
		}

	}

}

func (app *API) getSignal(c *gin.Context) {
	targetLink := c.Param(PARAM_USER_LINK)

	client, err := app.storage.GetSignal(targetLink)
	if err != nil {
		Error(c, http.StatusBadRequest, "Client signal not found")
		return
	}

	c.Header("Content-Type", "application/json")
	c.Writer.Write(client.SignalBytes)
	c.Status(http.StatusOK)
}

func (app *API) forwardSignal(c *gin.Context) {
	targetLink := c.Param(PARAM_USER_LINK)

	tokenPayload, err := getPlayload(c)
	if err != nil {
		Error(c, http.StatusForbidden, err.Error())
		return
	}

	clientId := tokenPayload.UserId

	targetSignal, err := app.storage.GetSignal(targetLink)
	if err != nil {
		Error(c, http.StatusBadRequest, "Target signal not found")
		return
	}

	var signal SignalData
	if err := c.ShouldBindBodyWithJSON(&signal); err != nil {
		Error(c, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	if !validSDP(signal.SDP) || !validCandidates(signal.Candidates) {
		Error(c, http.StatusBadRequest, "Invalid SDP or Candidates")
		return
	}

	if app.config.UnifiedMessage {

		message := &AuthMessage.Message[SignalData]{
			Type:    MESSAGE_TYPE,
			Content: signal,
		}

		messageBytes, _ := json.Marshal(message)

		request := &ServiceAuth.SendMessageRequest{
			ChannelId: targetSignal.ChannelId,
			ClientId:  targetSignal.ClientId,
			Message:   messageBytes,
		}

		if app.pulsar != nil {

			requestBytes, _ := json.Marshal(request)
			_, err := app.pulsar.SendMessage(targetSignal.ChannelId, requestBytes)
			if err != nil {
				Error(c, http.StatusInternalServerError, err)
				return
			}
		} else {

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			_, err := app.authClient.SendMessage(ctx, request)
			if err != nil {
				Error(c, http.StatusInternalServerError, err)
				return
			}
		}

	} else {

		signalChannel := app.signalChannels.Get(targetLink)
		if signalChannel != nil {

			signal.ClientId = clientId
			signalChannel <- signal

		} else if app.pulsar != nil {

			kafkerSignal := KafkerSignal{
				LinkCode: targetLink,
				Signal:   signal,
			}
			signalBytes, _ := json.Marshal(kafkerSignal)

			_, err := app.pulsar.SendMessage(targetSignal.ChannelId, signalBytes)
			if err != nil {
				Error(c, http.StatusInternalServerError, err)
				return
			}

		} else {
			Error(c, http.StatusNotFound, "")
			return
		}

	}

	c.Status(http.StatusOK)
}

// validSDP validates the Session Description Protocol (SDP) string.
func validSDP(rawSDP string) bool {
	sdp := strings.ReplaceAll(rawSDP, "\n", "&")

	query, err := url.ParseQuery(sdp)
	if err != nil {
		return false
	}

	for _, key := range []string{"v", "o", "s", "m", "c", "a"} {
		if !query.Has(key) {
			return false
		}
	}

	return true
}

// validCandidates validates the WebRTC candidates.
func validCandidates(candidates []Candidate) bool {
	for _, candidate := range candidates {
		if candidate.Candidate == nil || candidate.SdpMLineIndex == nil || candidate.SdpMid == nil {
			return false
		}
	}

	return true
}
