package clientendpoint

import (
	"fmt"
	"log"
	"net/http"
	AuthLifecycle "peergrine/jwtissuer/api/client-endpoint/auth-lifecycle"
	AppConfig "peergrine/jwtissuer/app-config"
	Storage "peergrine/jwtissuer/storage"
	Auth "peergrine/utils/auth"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	PARAM_USER_ID = "user_id" // 常數，用於上下文中的客戶端 ID 參數
)

type ClientEndpoint struct {
	server        *gin.Engine
	storage       *Storage.Storage
	tokenDuration AuthLifecycle.TokenDuration
}

func AtoD(str string) (time.Duration, error) {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}

	d := time.Duration(i) * time.Second

	return d, nil
}

// New 創建一個新的 API 實例。
// 參數:
//
//	config (*AppConfig.AppConfig): 應用程序配置。
//
// 返回值:
//
//	*API: 初始化的 API 實例。
func New(storage *Storage.Storage, config *AppConfig.AppConfig) (*ClientEndpoint, error) {

	bearerTokenDuration, err := AtoD(config.BearerTokenDuration)
	if err != nil {
		return nil, err
	}

	refreshTokenDuration, err := AtoD(config.RefreshTokenDuration)
	if err != nil {
		return nil, err
	}

	tokenDuration := AuthLifecycle.TokenDuration{
		Bearer:  bearerTokenDuration,
		Refresh: refreshTokenDuration,
	}

	server := gin.Default()

	app := &ClientEndpoint{
		server:        server,
		storage:       storage,
		tokenDuration: tokenDuration,
	}

	authLifecycle, err := AuthLifecycle.New(storage, tokenDuration)
	if err != nil {
		return nil, err
	}

	server.GET("/health", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})

	server.GET("/initialize", authLifecycle.InitializeAuth)
	server.POST("/refresh", app.RefreshToken)
	server.POST("/transfer", app.TransferToken)

	return app, nil
}

func (e *ClientEndpoint) Run(addr string) error {
	return e.server.Run(addr)
}

// Error 返回錯誤響應，並在內部伺服器錯誤時記錄錯誤。
// 參數:
//
//	c (*gin.Context): Gin 上下文對象，用於處理請求和響應。
//	statusCode (int): HTTP 狀態碼。
//	msg (any): 錯誤消息。
func Error(c *gin.Context, statusCode int, msg any) {
	if statusCode == http.StatusInternalServerError {
		log.Printf("[ERROR] Status %d: %v", statusCode, msg)
		c.Status(statusCode)
	} else {
		c.JSON(statusCode, gin.H{
			"error":   fmt.Sprintf("%v", msg),
			"message": http.StatusText(statusCode),
			"status":  statusCode,
		})
	}
	c.Abort()
}

// Refresh 根據提供的刷新令牌生成新的 Bearer token。如果刷新令牌無效或過期，返回未授權錯誤。
// 參數:
//
//	c (*gin.Context): Gin 上下文對象，用於處理請求和響應。
func (app *ClientEndpoint) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	userId, exists := app.storage.GetUserIdFromRefreshToken(authHeader)

	if !exists {
		Error(c, http.StatusUnauthorized, "Refresh token is invalid or expired")
		return
	}

	currentTime := time.Now()
	iat := currentTime.Unix()
	exp := currentTime.Add(app.tokenDuration.Bearer).Unix()

	serviceId := app.storage.ServiceId
	secret, err := app.storage.GetSecret(serviceId)
	if err != nil {
		Error(c, http.StatusInternalServerError, err)
	}

	bearerToken, err := Auth.GenerateBearerToken(serviceId, userId, secret, iat, exp)
	if err != nil {
		Error(c, http.StatusInternalServerError, "Failed to generate new access token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": bearerToken,
		"expires_at":   exp,
	})
}

func (app *ClientEndpoint) TransferToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	userId, exists := app.storage.GetUserIdFromRefreshToken(authHeader)

	if !exists {
		Error(c, http.StatusUnauthorized, "Refresh token is invalid or expired")
		return
	}

	app.storage.DeleteToken(authHeader)

	currentTime := time.Now()
	iat := currentTime.Unix()
	exp := currentTime.Add(app.tokenDuration.Bearer).Unix()

	serviceId := app.storage.ServiceId
	secret, err := app.storage.GetSecret(serviceId)
	if err != nil {
		Error(c, http.StatusInternalServerError, err)
	}

	refreshToken, err := Auth.GenerateRefreshToken(serviceId, userId, secret, currentTime)
	if err != nil {
		Error(c, http.StatusInternalServerError, "Failed to generate new refresh token")
		return
	}

	bearerToken, err := Auth.GenerateBearerToken(serviceId, userId, secret, iat, exp)
	if err != nil {
		Error(c, http.StatusInternalServerError, "Failed to generate new access token")
		return
	}

	app.storage.SaveToken(refreshToken, userId, app.tokenDuration.Refresh)

	c.JSON(http.StatusOK, gin.H{
		"refresh_token": refreshToken,
		"access_token":  bearerToken,
		"expires_at":    exp,
	})
}
