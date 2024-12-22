package rtcbridgeapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	ServiceAuth "peergrine/grpc/serviceauth"
	AppConfig "peergrine/rtc-bridge/app-config"
	Storage "peergrine/rtc-bridge/storage"
	Auth "peergrine/utils/auth"
	GenericChannels "peergrine/utils/generic-channels"
	Pulsar "peergrine/utils/pulsar"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	PARAM_USER_ID   = "user_id"   // 常數，用於上下文中的客戶端 ID 參數
	PARAM_USER_LINK = "user_link" // 常數，用於路徑中的客戶端連結參數
	TOKEN_PARLOAD   = "payload"
)

type API struct {
	config             *AppConfig.AppConfig
	storage            *Storage.Storage
	authConnection     *grpc.ClientConn
	authClient         ServiceAuth.ServiceAuthClient
	signalChannels     GenericChannels.Channels[SignalData]
	pulsar             *Pulsar.Client
	server             *http.Server
	stopListenMessages context.CancelFunc
}

// New 創建新的 API 實例並設置相關的路由和中介軟體
func New(config *AppConfig.AppConfig, storage *Storage.Storage, pulsar *Pulsar.Client) (*API, error) {

	app := &API{
		config:         config,
		storage:        storage,
		signalChannels: GenericChannels.New[SignalData](),
		pulsar:         pulsar,
	}

	if config.AuthAddr != "" {

		conn, err := grpc.NewClient(config.AuthAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}

		app.authConnection = conn
		app.authClient = ServiceAuth.NewServiceAuthClient(conn)
	}

	if pulsar != nil && !config.UnifiedMessage {

		ctx, cancel := context.WithCancel(context.Background())
		app.stopListenMessages = cancel
		go app.listenPulsarMessages(ctx)

	}

	router := gin.Default()

	signalRoutes := router.Group("/", app.authRequired)

	{
		signalRoutes.GET(":"+PARAM_USER_LINK, app.getSignal)      // 獲取信號
		signalRoutes.POST(":"+PARAM_USER_LINK, app.forwardSignal) // 轉發信號
		signalRoutes.POST("", app.setSignal)                      // 設置信號
	}

	app.server = &http.Server{
		Addr:    config.Addr,
		Handler: router,
	}

	return app, nil
}

func (app *API) Run() error {
	return app.server.ListenAndServe()
}

func (app *API) Close() {
	if app.stopListenMessages != nil {
		app.stopListenMessages()
	}
	app.signalChannels.Close()
	if app.authConnection != nil {
		app.authConnection.Close()
		app.storage.Close()
	}
}

func (app *API) listenPulsarMessages(ctx context.Context) {

	for msg := range app.pulsar.ListenMessages(ctx, 10) {

		var kafkerSignal KafkerSignal
		err := json.Unmarshal(msg, &kafkerSignal)
		if err == nil {

			linkCode := kafkerSignal.LinkCode

			signalChannel := app.signalChannels.Get(linkCode)

			if signalChannel != nil {
				if kafkerSignal.Signal.ClientId == "" {
					app.signalChannels.Del(linkCode)
				} else {
					signalChannel <- kafkerSignal.Signal
				}
			}

		}

	}

}

// Error 處理錯誤消息並根據狀態碼響應
func Error(c *gin.Context, statusCode int, msg any) {

	if statusCode == http.StatusInternalServerError {
		log.Println(msg) // 記錄內部伺服器錯誤的錯誤消息
		c.Status(statusCode)
	} else {
		c.JSON(statusCode, msg) // 返回 JSON 格式的錯誤消息
	}
	c.Abort() // 中止請求
}

// AuthRequired 中介軟體進行授權，檢查 Authorization 標頭並驗證令牌
func (app *API) authRequired(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		Error(c, http.StatusUnauthorized, "Authorization header is missing or formatted incorrectly. Expected format: 'Bearer <token>'")
		return
	}

	bearerToken := authHeader[7:]

	cacheTokenPayload := app.storage.GetTokenCache(bearerToken)
	if cacheTokenPayload != nil {
		c.Set(TOKEN_PARLOAD, *cacheTokenPayload)
	} else {

		if app.authConnection != nil {

			req := &ServiceAuth.AccessTokenRequest{
				AccessToken: bearerToken,
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
			defer cancel()

			res, err := app.authClient.VerifyAccessToken(ctx, req)

			if err != nil {
				log.Println(err)
				Error(c, http.StatusUnauthorized, "a:Token is invalid or has expired. Please provide a valid token.")
				return
			}

			tokenPayload := Auth.TokenPayload{
				Token:     bearerToken,
				Iss:       res.Iss,
				Iat:       res.Iat,
				Exp:       res.Exp,
				UserId:    res.UserId,
				ChannelId: res.ChannelId,
			}

			app.storage.SetTokenCache(bearerToken, tokenPayload)

			c.Set(TOKEN_PARLOAD, tokenPayload)

		} else {

			iss, err := Auth.ExtractIssuerFromToken(bearerToken)
			if err != nil {
				Error(c, http.StatusUnauthorized, "Failed to extract issuer from token. Token may be malformed or invalid.")
				return
			}

			secret, err := app.storage.GetSecret(iss)
			if err != nil {
				Error(c, http.StatusInternalServerError, err)
				return
			}

			claims, err := Auth.DecodeToken(bearerToken, secret)
			if err != nil {
				Error(c, http.StatusUnauthorized, "l:Token is invalid or has expired. Please provide a valid token.")
				return
			}

			tokenPayload := Auth.Claims2TokenPayload(bearerToken, claims)

			app.storage.SetTokenCache(bearerToken, tokenPayload)
			c.Set(TOKEN_PARLOAD, tokenPayload)

		}

	}

	c.Next()
}
