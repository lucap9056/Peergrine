package authlifecycle

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	ConnMap "peergrine/jwtissuer/api/conn-map"
	Messages "peergrine/jwtissuer/client-messages"
	Storage "peergrine/jwtissuer/storage"
	Auth "peergrine/utils/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type TokenDuration struct {
	Bearer  time.Duration
	Refresh time.Duration
}

type Manager struct {
	connMap       *ConnMap.ConnMap
	mutex         *sync.RWMutex
	storage       *Storage.Storage
	tokenDuration TokenDuration
	channelId     int32
}

// New 創建一個新的 WSS 管理器實例。
// 參數:
//
//	config (*AppConfig.AppConfig): 應用程序配置。
//
// 返回值:
//
//	(*Manager, error): 初始化的 Manager 實例和錯誤信息（如果有）。
func New(storage *Storage.Storage, connMap *ConnMap.ConnMap, tokenDuration TokenDuration, channelId int32) (*Manager, error) {
	wss := &Manager{
		connMap:       connMap,
		mutex:         new(sync.RWMutex),
		storage:       storage,
		tokenDuration: tokenDuration,
		channelId:     channelId,
	}

	return wss, nil
}

// InitializeAuth 處理 WebSocket 連接。生成令牌，將客戶端連接升級為 WebSocket，並處理傳入消息。
// 參數:
//
//	c (*gin.Context): Gin 上下文對象，用於處理請求和響應。
func (wss *Manager) InitializeAuth(c *gin.Context) {

	serviceId := wss.storage.ServiceId
	secret, err := wss.storage.GetSecret(serviceId)
	if err != nil {
		c.Error(err)
		return
	}

	id := uuid.New().String()

	currentTime := time.Now()

	refreshToken, err := Auth.GenerateRefreshToken(serviceId, id, wss.channelId, secret, currentTime)
	if err != nil {
		c.Error(err)
		return
	}

	iat := currentTime.Unix()
	exp := currentTime.Add(wss.tokenDuration.Bearer).Unix()

	bearerToken, err := Auth.GenerateBearerToken(serviceId, id, wss.channelId, secret, iat, exp)
	if err != nil {
		c.Error(err)
		return
	}

	wss.storage.SaveToken(refreshToken, id, wss.tokenDuration.Refresh)
	defer wss.storage.DeleteToken(refreshToken)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.Error(err)
		return
	}

	wss.connMap.Set(id, conn)

	authorizeMessage := Messages.Authorization(refreshToken, bearerToken, exp)

	conn.WriteJSON(authorizeMessage)

	for {
		msgType, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			conn.Close()
			wss.removeClient(id)
			return
		}

		switch msgType {
		case websocket.CloseMessage:
			conn.Close()
			wss.removeClient(id)
			return
		}
	}
}

// removeClient 從客戶端列表中移除並關閉指定的 WebSocket 連接。
// 參數:
//
//	id (string): 客戶端的唯一標識符。
func (wss *Manager) removeClient(id string) {
	wss.connMap.Del(id)
}

// HasClient 檢查客戶端是否在連接列表中。
// 參數:
//
//	id (string): 客戶端的唯一標識符。
//
// 返回值:
//
//	bool: 如果客戶端存在於列表中，則返回 true，否則返回 false。
func (wss *Manager) HasClient(id string) bool {
	_, exists := wss.connMap.Get(id)
	return exists
}
