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
	Storage "peergrine/rtc-bridge/storage"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Constants
const LINK_CODE_LETTER_BYTES = "abcdefghijkmnopqrstuvwxyzABCDEFHJKLMNPQRSTUVWXYZ0123456789"
const LINK_CODE_DURATION = 5 * time.Minute

// generateUniqueLinkCode generates a unique link code for signal identification.
// It attempts to generate a unique code up to a maximum number of times, and checks
// whether the generated code already exists in the signal storage.
// Returns:
//   - string: The unique link code if successful.
//   - error: Returns an error if unable to generate a unique code after the maximum attempts.
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

// setSignal receives and storages the signal data, and waits for an incoming signal for a limited time.
// The request is processed to storage a unique signal, with a link code for future retrieval.
// Parameters:
//   - c (gin.Context): The Gin context for the incoming HTTP request.
//
// The function returns an HTTP response with a link code and expiration time, or an error message.
func (app *API) setSignal(c *gin.Context) {
	clientIdValue, exists := c.Get(PARAM_USER_ID)
	if !exists {
		Error(c, http.StatusForbidden, "Client ID not found in context")
		return
	}
	clientId, ok := clientIdValue.(string)
	if !ok {
		Error(c, http.StatusForbidden, "Invalid Client ID format")
		return
	}

	var signal SignalData
	if err := c.ShouldBindJSON(&signal); err != nil {
		Error(c, http.StatusBadRequest, "Invalid JSON format")
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
	signalBytes, err := json.Marshal(signal)
	if err != nil {

		Error(c, http.StatusInternalServerError, fmt.Sprintf("failed to marshal signal: %v", err))
		return
	}

	expiresAt := time.Now().Add(LINK_CODE_DURATION).Unix()

	clientSignal := Storage.NewSignal(clientId, linkCode, signalBytes, expiresAt)

	if err := app.storage.SetSignal(clientSignal); err != nil {

		Error(c, http.StatusInternalServerError, fmt.Sprintf("failed to storage client signal: %v", err))
		return
	}

	result := LinkCode{
		LinkCode:  linkCode,
		ExpiresAt: expiresAt,
	}
	c.JSON(http.StatusOK, result)
	c.Writer.Flush()

	awaitSignal := app.signalChannels.Add(linkCode)
	defer app.signalChannels.Del(linkCode)

	ctx, cancel := context.WithTimeout(context.Background(), LINK_CODE_DURATION)
	defer cancel()

	select {
	case targetSignal, ok := <-awaitSignal:
		if ok {
			c.JSON(http.StatusOK, targetSignal)
		}
	case <-ctx.Done():
		Error(c, http.StatusRequestTimeout, "Request timed out")
	}

}

// getSignal retrieves signal data from the storage based on the provided link code.
// The client sends the link code in the URL, and this function retrieves and returns the
// associated signal data.
// Parameters:
//   - c (gin.Context): The Gin context for the incoming HTTP request.
//
// The function returns the storaged signal data in JSON format, or an error message.
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

// forwardSignal forwards the signal data to the specified link code.
// This function retrieves the signal by link code and forwards the SDP and candidates to
// another client. If the channel is available, it sends the signal through the channel,
// otherwise it forwards via Kafka if the channel is unavailable.
// Parameters:
//   - c (gin.Context): The Gin context for the incoming HTTP request.
//
// The function sends a forwarded signal to the appropriate client and returns an HTTP status.
func (app *API) forwardSignal(c *gin.Context) {
	targetLink := c.Param(PARAM_USER_LINK)

	clientId, exists := c.Get(PARAM_USER_ID)
	if !exists {
		Error(c, http.StatusForbidden, "Client ID not found in context")
		return
	}

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

	signalChannel := app.signalChannels.Get(targetLink)
	if signalChannel != nil {

		signal.ClientId = clientId.(string)
		signalChannel <- signal

		c.Status(http.StatusOK)
		return
	}

	kafkerSignal := KafkerSignal{
		LinkCode: targetLink,
		Signal:   signal,
	}

	signalBytes, err := json.Marshal(kafkerSignal)
	if err != nil {

		Error(c, http.StatusBadRequest, "Invalid JSON format")

	} else if app.kafka != nil {

		_, _, err := app.kafka.Producer.SendMessage(app.config.KafkaTopic, signalBytes, targetSignal.ChannelId)
		if err != nil {
			Error(c, http.StatusInternalServerError, err)
		}

	}
}

// removeSignal removes a signal based on the provided link code.
// If the link code is found and the requesting client owns the signal, it is deleted from the storage.
// If the signal is in an active channel, it closes the channel; otherwise, it sends a removal request via Kafka.
// Parameters:
//   - c (gin.Context): The Gin context for the incoming HTTP request.
//
// The function deletes the signal from the storage and returns an HTTP status.
func (app *API) removeSignal(c *gin.Context) {
	linkCode := c.Param(PARAM_USER_LINK)
	clientId, exists := c.Get(PARAM_USER_ID)
	if !exists {
		Error(c, http.StatusForbidden, "Client ID not found in context")
		return
	}

	signal, err := app.storage.GetSignal(linkCode)
	if err != nil {
		Error(c, http.StatusBadRequest, "Client signal not found")
		return
	}

	if clientId != signal.ClientId {
		Error(c, http.StatusUnauthorized, "Client is not signal owner")
		return
	}

	signalChannel := app.signalChannels.Get(linkCode)
	if signalChannel != nil {
		app.signalChannels.Del(linkCode)
	} else if app.kafka != nil {
		kafkerSignal := KafkerSignal{
			LinkCode: linkCode,
		}

		signalBytes, err := json.Marshal(kafkerSignal)
		if err != nil {
			Error(c, http.StatusInternalServerError, err)
		} else {

			_, _, err := app.kafka.SendMessage(app.config.KafkaTopic, signalBytes, signal.ChannelId)
			if err != nil {
				Error(c, http.StatusInternalServerError, err)
			}

		}

	}

	app.storage.RemoveSignal(linkCode)

	c.Status(http.StatusOK)
}

// validSDP validates the Session Description Protocol (SDP) string.
// The function checks for the required SDP fields (e.g., "v", "o", "s", etc.) in the input string
// and returns a boolean indicating whether the SDP is valid.
// Parameters:
//   - rawSDP (string): The raw SDP string to be validated.
//
// Returns:
//   - bool: True if the SDP is valid, otherwise false.
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
// The function checks each candidate's fields (e.g., Candidate, SdpMLineIndex, and SdpMid)
// to ensure they are not nil and properly structured.
// Parameters:
//   - candidates ([]Candidate): A slice of Candidate structs to be validated.
//
// Returns:
//   - bool: True if all candidates are valid, otherwise false.
func validCandidates(candidates []Candidate) bool {
	for _, candidate := range candidates {
		if candidate.Candidate == nil || candidate.SdpMLineIndex == nil || candidate.SdpMid == nil {
			return false
		}
	}

	return true
}
