package msgbridgeapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	ServiceAuth "peergrine/grpc/serviceauth"
	AppConfig "peergrine/msg-bridge/app-config"
	Storage "peergrine/msg-bridge/storage"
	Auth "peergrine/utils/auth"
	GenericChannels "peergrine/utils/generic-channels"
	Kafka "peergrine/utils/kafka"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	PARAM_USER_ID   = "user_id"   // Constant for user ID parameter
	PARAM_LINK_CODE = "link_code" // Constant for user link parameter
)

type Server struct {
	config          *AppConfig.AppConfig
	storage         *Storage.Storage
	authConnection  *grpc.ClientConn
	authClient      ServiceAuth.ServiceauthClient
	messageChannels GenericChannels.Channels[[]byte]
	kafka           *Kafka.Client
	kafkaChannelId  int32
	server          *http.Server
}

// New creates and initializes a new Server instance with configuration, storage, and Kafka client.
// It also sets up the gRPC authentication client if the auth service address is provided.
func New(config *AppConfig.AppConfig, storage *Storage.Storage, kafka *Kafka.Client, kafkaChannelId int32) (*Server, error) {

	app := &Server{
		config:          config,
		storage:         storage,
		messageChannels: GenericChannels.New[[]byte](),
		kafka:           kafka,
		kafkaChannelId:  kafkaChannelId,
	}

	if config.AuthAddr != "" {

		conn, err := grpc.NewClient(config.AuthAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		app.authConnection = conn
		app.authClient = ServiceAuth.NewServiceauthClient(conn)
	}

	if kafka != nil {
		go app.listenKafkerMessage()
	}

	router := gin.Default()

	// Define message routes with authentication middleware
	messageRoutes := router.Group("/", app.authRequired)
	{
		messageRoutes.POST("/session", app.postPublicKey)                     // POST: Save RSA public key and create an link code
		messageRoutes.POST("/session/:"+PARAM_LINK_CODE, app.getClient)       // GET: Retrieve specific client's public key using link code
		messageRoutes.DELETE("/session/:"+PARAM_LINK_CODE, app.removeSession) // DELETE: Remove link code
		messageRoutes.GET("/messages", app.getSSE)                            // GET: Establish SSE to receive messages
		messageRoutes.POST("/messages/:"+PARAM_USER_ID, app.postMessage)      // POST: Send an encrypted message to a specific client
	}

	app.server = &http.Server{
		Addr:    config.Addr,
		Handler: router,
	}

	return app, nil
}

// Run starts the HTTP server and listens for incoming requests.
func (app *Server) Run() error {
	return app.server.ListenAndServe()
}

// Close gracefully closes the gRPC connection and any other resources associated with the server.
func (app *Server) Close() {
	app.messageChannels.Close()
	if app.authConnection != nil {
		app.authConnection.Close()
		app.storage.Close()
	}
}

func (app *Server) listenKafkerMessage() {
	handler := make(chan []byte)
	done := make(chan interface{})

	app.kafka.ConsumeMessages(app.config.KafkaTopic, app.kafkaChannelId, sarama.OffsetNewest, handler, done)

	for {
		msg := <-handler

		var kafkaMessage KafkaMessage
		err := json.Unmarshal(msg, &kafkaMessage)
		if err == nil {
			clientId := kafkaMessage.ClientId

			messageChannel := app.messageChannels.Get(clientId)

			if messageChannel != nil {
				messageChannel <- kafkaMessage.Message
			}

		}
		done <- nil
	}
}

// Error handles errors by logging internal server errors and sending appropriate HTTP responses.
// It also aborts the request to ensure no further processing.
func Error(c *gin.Context, statusCode int, msg any) {
	if statusCode == http.StatusInternalServerError {
		log.Println(msg) // Log internal server errors
		c.Status(statusCode)
	} else {
		c.JSON(statusCode, msg) // Return JSON formatted error message
	}
	c.Abort() // Abort the request to stop further handling
}

// authRequired is middleware that performs authorization by checking the Authorization header for a Bearer token.
// It verifies the token using the auth service or local validation and sets the user ID in the context if successful.
func (app *Server) authRequired(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		Error(c, http.StatusUnauthorized, "Authorization header is missing or formatted incorrectly. Expected format: 'Bearer <token>'")
		return
	}

	bearerToken := authHeader[7:]

	// Check token in cache
	cacheTokenData := app.storage.GetTokenCache(bearerToken)
	if cacheTokenData != nil {
		c.Set(PARAM_USER_ID, cacheTokenData.UserId)
	} else {
		// If not cached, verify token with auth service
		if app.authConnection != nil {
			req := &ServiceAuth.AccessTokenRequest{
				AccessToken: bearerToken,
			}
			res, err := app.authClient.VerifyAccessToken(context.Background(), req)
			if err != nil {
				log.Println(err)
				Error(c, http.StatusUnauthorized, "Token is invalid or has expired. Please provide a valid token.")
				return
			}

			tokenData := Auth.TokenData{
				Token:  bearerToken,
				Iss:    res.Iss,
				Iat:    res.Iat,
				Exp:    res.Exp,
				UserId: res.UserId,
			}

			app.storage.SetTokenCache(bearerToken, tokenData)
			c.Set(PARAM_USER_ID, res.UserId)

		} else {
			// Perform local token validation if no auth service is available
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
				Error(c, http.StatusUnauthorized, "Token is invalid or has expired. Please provide a valid token.")
				return
			}

			iat, _ := (*claims)["iat"].(float64)
			exp, _ := (*claims)["exp"].(float64)

			tokenData := Auth.TokenData{
				Token:  bearerToken,
				Iss:    (*claims)["iss"].(string),
				Iat:    int64(iat),
				Exp:    int64(exp),
				UserId: (*claims)["user_id"].(string),
			}

			app.storage.SetTokenCache(bearerToken, tokenData)
			c.Set(PARAM_USER_ID, tokenData.UserId)
		}
	}

	c.Next() // Continue processing the request if authentication succeeds
}
