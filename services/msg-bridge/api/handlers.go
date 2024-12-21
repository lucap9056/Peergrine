package msgbridgeapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	ServiceAuth "peergrine/grpc/serviceauth"
	AuthMessage "peergrine/jwtissuer/client-messages"
	Storage "peergrine/msg-bridge/storage"
	Auth "peergrine/utils/auth"
	"time"

	"github.com/gin-gonic/gin"
)

// Constants
const LINK_CODE_LETTER_BYTES = "abcdefghijkmnopqrstuvwxyzABCDEFHJKLMNPQRSTUVWXYZ0123456789"
const LINK_CODE_DURATION = 5 * time.Minute

const MESSAGE_TYPE = "message-relay"

// generateUniqueLinkCode generates a unique link code for signal identification.
// It attempts to generate a unique code up to a maximum number of times, and checks
// whether the generated code already exists in the signal storage.
// Returns:
//   - string: The unique link code if successful.
//   - error: Returns an error if unable to generate a unique code after the maximum attempts.
func (app *Server) generateUniqueLinkCode() (string, error) {
	const maxAttempts = 5
	const linkCodeLength = 8

	letterLen := len(LINK_CODE_LETTER_BYTES)

	for attempts := 0; attempts < maxAttempts; attempts++ {
		b := make([]byte, linkCodeLength)
		for i := range b {
			randNum, err := rand.Int(rand.Reader, big.NewInt(int64(letterLen)))
			if err != nil {
				return "", fmt.Errorf("failed to generate random number for link code: %v", err)
			}
			b[i] = LINK_CODE_LETTER_BYTES[randNum.Int64()]
		}

		linkCode := string(b)

		exist, err := app.storage.ClientSessionExists(linkCode)
		if err != nil {
			return "", fmt.Errorf("error checking link code existence in storage: %v", err)
		}
		if !exist {
			return linkCode, nil
		}
	}

	return "", fmt.Errorf("exceeded maximum attempts (%d) to generate a unique link code", maxAttempts)
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

func (app *Server) getChannelId(payload *Auth.TokenPayload) (unifiedMessage bool, channelId int32) {
	if app.config.UnifiedMessage {
		return true, payload.ChannelId
	}
	return false, app.kafkaChannelId
}

func (app *Server) postPublicKey(c *gin.Context) {
	tokenPayload, err := getPlayload(c)
	if err != nil {
		Error(c, http.StatusForbidden, err.Error())
		return
	}

	bodyBytes, err := c.GetRawData()
	if err != nil {
		Error(c, http.StatusBadRequest, "Failed to read request body")
		return
	}

	linkCode, err := app.generateUniqueLinkCode()
	if err != nil {
		Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to generate a unique link code: %v", err))
		return
	}

	expiresAt := time.Now().Add(LINK_CODE_DURATION).Unix()

	session := SessionData{
		ClientId:  tokenPayload.UserId,
		PublicKey: string(bodyBytes),
	}

	sessionBytes, err := json.Marshal(session)
	if err != nil {
		Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to marshal session data: %v", err))
		return
	}

	clientSession := Storage.ClientSession{
		LinkCode:     linkCode,
		ClientId:     tokenPayload.UserId,
		ChannelId:    tokenPayload.ChannelId,
		SessionBytes: sessionBytes,
		ExpiresAt:    expiresAt,
	}

	if err := app.storage.SetClientSession(clientSession); err != nil {
		Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to store client session: %v", err))
		return
	}

	result := LinkCode{
		LinkCode:  linkCode,
		ExpiresAt: expiresAt,
	}

	c.JSON(http.StatusOK, result)
}

func (app *Server) listenMessage(c *gin.Context) {
	tokenPayload, err := getPlayload(c)
	if err != nil {
		Error(c, http.StatusForbidden, err.Error())
		return
	}

	clientId := tokenPayload.UserId

	unifiedMessage, channelId := app.getChannelId(tokenPayload)

	if err := app.storage.SetClientChannel(clientId, channelId); err != nil {
		Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to set client channel in storage: %v", err))
		return
	}
	defer app.storage.RemoveClientChannel(clientId)

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	c.String(http.StatusOK, "event: connected")
	c.Writer.Flush()

	closeNotify := c.Writer.CloseNotify()

	if unifiedMessage {
		<-closeNotify
	} else {

		messageChannnel := app.messageChannels.Add(clientId)
		defer app.messageChannels.Del(clientId)
		for {
			select {
			case <-closeNotify:
				return
			case message, ok := <-messageChannnel:
				if ok {
					c.Writer.Write(message)
					c.Writer.Flush()
				} else {
					return
				}
			}
		}

	}

}

func (app *Server) getClient(c *gin.Context) {
	tokenPayload, err := getPlayload(c)
	if err != nil {
		Error(c, http.StatusForbidden, err.Error())
		return
	}

	targetLink := c.Param(PARAM_LINK_CODE)

	target, err := app.storage.GetClientSession(targetLink)
	if err != nil {
		Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve client session for link: %s. Error: %v", targetLink, err))
		return
	}

	bodyBytes, err := c.GetRawData()
	if err != nil {
		Error(c, http.StatusBadRequest, "Failed to read request body")
		return
	}

	c.Header("Content-Type", "application/json")
	c.Writer.Write(target.SessionBytes)
	c.Status(http.StatusOK)

	targetId := target.ClientId

	data := SessionData{
		ClientId:  tokenPayload.UserId,
		PublicKey: string(bodyBytes),
	}

	if app.config.UnifiedMessage {

		mesage := AuthMessage.Message[SessionData]{
			Type:    MESSAGE_TYPE,
			Content: data,
		}

		messageBytes, _ := json.Marshal(mesage)

		request := &ServiceAuth.SendMessageRequest{
			ChannelId: target.ChannelId,
			ClientId:  targetId,
			Message:   messageBytes,
		}

		if app.kafka != nil {

			requestBytes, _ := json.Marshal(request)

			_, _, err := app.kafka.SendMessage(app.config.KafkaTopic, requestBytes, target.ChannelId)
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

		dataBytes, err := json.Marshal(data)
		if err != nil {
			Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to marshal message data: %v", err))
			return
		}

		sessionStr := fmt.Sprintf("event: append_user\n\ndata: %s\n\n", dataBytes)

		sessionBytes := []byte(sessionStr)

		messageChannel := app.messageChannels.Get(targetId)
		if messageChannel != nil {
			messageChannel <- sessionBytes
		} else if app.kafka != nil {
			channelId, err := app.storage.GetClientChannel(targetId)
			if err != nil {
				Error(c, http.StatusNotFound, fmt.Sprintf("Client channel not found for target ID: %s. Error: %v", targetId, err))
				return
			}

			kafkaMessage := ForawrdMessage{
				ClientId: targetId,
				Content:  sessionBytes,
			}

			kafkaMessageBytes, _ := json.Marshal(kafkaMessage)

			_, _, err = app.kafka.SendMessage(app.config.KafkaTopic, kafkaMessageBytes, channelId)
			if err != nil {
				Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to send message via Kafka for target ID: %s. Error: %v", targetId, err))
				return
			}

		} else {
			Error(c, http.StatusNotFound, "")
			return
		}

	}

	c.Status(http.StatusOK)

}

func (app *Server) postMessage(c *gin.Context) {
	targetId := c.Param(PARAM_USER_ID)

	tokenPayload, err := getPlayload(c)
	if err != nil {
		Error(c, http.StatusForbidden, err.Error())
		return
	}

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		Error(c, http.StatusBadRequest, "Failed to read request body")
		return
	}

	data := MessageData{
		SenderId: tokenPayload.UserId,
		Message:  string(bodyBytes),
	}

	dataJson, err := json.Marshal(data)
	if err != nil {
		Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to marshal message data: %v", err))
		return
	}

	if app.config.UnifiedMessage {

		message := AuthMessage.Message[MessageData]{
			Type:    MESSAGE_TYPE,
			Content: data,
		}

		messageBytes, _ := json.Marshal(message)

		request := &ServiceAuth.SendMessageRequest{
			ChannelId: 0,
			ClientId:  targetId,
			Message:   messageBytes,
		}

		if app.kafka != nil {
			channelId, err := app.storage.GetClientChannel(targetId)
			if err != nil {
				Error(c, http.StatusNotFound, fmt.Sprintf("Client channel not found for target ID: %s. Error: %v", targetId, err))
				return
			}

			request.ChannelId = channelId

			requestBytes, _ := json.Marshal(request)

			_, _, err = app.kafka.SendMessage(app.config.KafkaTopic, requestBytes, channelId)
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

		messageStr := fmt.Sprintf("event: message\n\ndata: %s\n\n", dataJson)
		messageBytes := []byte(messageStr)

		messageChannel := app.messageChannels.Get(targetId)
		if messageChannel != nil {
			messageChannel <- messageBytes
		} else if app.kafka != nil {

			channelId, err := app.storage.GetClientChannel(targetId)
			if err != nil {
				Error(c, http.StatusNotFound, fmt.Sprintf("Client channel not found for target ID: %s. Error: %v", targetId, err))
				return
			}

			kafkaMessage := ForawrdMessage{
				ClientId: targetId,
				Content:  messageBytes,
			}

			kafkaMessageBytes, _ := json.Marshal(kafkaMessage)

			_, _, err = app.kafka.SendMessage(app.config.KafkaTopic, kafkaMessageBytes, channelId)
			if err != nil {
				Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to send message via Kafka for target ID: %s. Error: %v", targetId, err))
				return
			}

		} else {
			Error(c, http.StatusNotFound, "")
			return
		}

	}

	c.Status(http.StatusOK)
}

func (app *Server) removeSession(c *gin.Context) {
	linkCode := c.Param(PARAM_LINK_CODE)
	tokenPayload, err := getPlayload(c)
	if err != nil {
		Error(c, http.StatusForbidden, err.Error())
		return
	}

	key, err := app.storage.GetClientSession(linkCode)
	if err != nil {
		Error(c, http.StatusBadRequest, "Client signal not found")
		return
	}

	if tokenPayload.UserId != key.ClientId {
		Error(c, http.StatusUnauthorized, "Client is not signal owner")
		return
	}

	app.storage.RemoveClientSession(linkCode)
	c.Status(http.StatusOK)
}
