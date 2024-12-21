package serviceendpoint

import (
	"context"
	"encoding/json"
	"net"
	ServiceAuth "peergrine/grpc/serviceauth"
	ConnMap "peergrine/jwtissuer/api/conn-map"
	AppConfig "peergrine/jwtissuer/app-config"
	Storage "peergrine/jwtissuer/storage"
	Auth "peergrine/utils/auth"
	Kafka "peergrine/utils/kafka"

	"github.com/IBM/sarama"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

func (s *App) VerifyAccessToken(ctx context.Context, req *ServiceAuth.AccessTokenRequest) (*ServiceAuth.TokenResponse, error) {

	iss, err := Auth.ExtractIssuerFromToken(req.AccessToken)
	if err != nil {
		return nil, err
	}

	secret, err := s.storage.GetSecret(iss)
	if err != nil {
		return nil, err
	}

	claims, err := Auth.DecodeToken(req.AccessToken, secret)
	if err != nil {
		return nil, err
	}

	iat, _ := (*claims)["iat"].(float64)
	exp, _ := (*claims)["exp"].(float64)
	channelId, _ := (*claims)["channel_id"].(float64)

	res := ServiceAuth.TokenResponse{
		Iss:       (*claims)["iss"].(string),
		Iat:       int64(iat),
		Exp:       int64(exp),
		UserId:    (*claims)["user_id"].(string),
		ChannelId: int32(channelId),
	}

	return &res, nil
}

func (s *App) SendMessage(ctx context.Context, req *ServiceAuth.SendMessageRequest) (*ServiceAuth.SendMessageResponse, error) {

	if s.kafkaChannelId == req.ChannelId {
		conn, ok := s.connMap.Get(req.ClientId)
		if ok {
			err := conn.WriteMessage(websocket.TextMessage, req.Message)
			if err != nil {
				return nil, err
			}

		}
	} else {
		message, _ := json.Marshal(req)

		_, _, err := s.kafka.SendMessage(s.config.KafkaTopic, message, req.ChannelId)
		if err != nil {
			return nil, err
		}
	}

	return &ServiceAuth.SendMessageResponse{
		Success: true,
	}, nil
}

type App struct {
	ServiceAuth.UnimplementedServiceAuthServer
	server         *grpc.Server
	config         *AppConfig.AppConfig
	storage        *Storage.Storage
	connMap        *ConnMap.ConnMap
	kafka          *Kafka.Client
	kafkaChannelId int32
}

func New(storage *Storage.Storage, config *AppConfig.AppConfig, connMap *ConnMap.ConnMap, kafka *Kafka.Client, kafkaChannelId int32) *App {
	server := grpc.NewServer()

	app := &App{
		server:         server,
		config:         config,
		storage:        storage,
		connMap:        connMap,
		kafka:          kafka,
		kafkaChannelId: kafkaChannelId,
	}

	ServiceAuth.RegisterServiceAuthServer(server, app)

	if kafka != nil {
		go app.listenKafkerMessage()
	}

	return app
}

func (e *App) Run(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return e.server.Serve(listener)
}

func (e *App) Close() {
	e.storage.Close()
	e.server.Stop()
}

func (app *App) listenKafkerMessage() {
	handler := make(chan []byte)
	done := make(chan interface{})

	app.kafka.ConsumeMessages(app.config.KafkaTopic, app.kafkaChannelId, sarama.OffsetNewest, handler, done)

	for {
		msg := <-handler

		var req ServiceAuth.SendMessageRequest
		err := json.Unmarshal(msg, &req)
		if err == nil {
			clientId := req.ClientId

			conn, ok := app.connMap.Get(clientId)
			if ok {
				conn.WriteMessage(websocket.TextMessage, req.Message)
			}

		}
		done <- nil
	}
}
