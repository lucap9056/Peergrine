package serviceendpoint

import (
	"context"
	"encoding/json"
	"net"
	ServiceAuth "peergrine/grpc/serviceauth"
	ServiceUnifiedMessage "peergrine/grpc/unifiedmessage"
	ConnMap "peergrine/jwtissuer/api/conn-map"
	AppConfig "peergrine/jwtissuer/app-config"
	Storage "peergrine/jwtissuer/storage"
	Auth "peergrine/utils/auth"
	Pulsar "peergrine/utils/pulsar"

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

	res := ServiceAuth.TokenResponse{
		Iss:       (*claims)["iss"].(string),
		Iat:       int64(iat),
		Exp:       int64(exp),
		UserId:    (*claims)["user_id"].(string),
		ChannelId: (*claims)["channel_id"].(string),
	}

	return &res, nil
}

func (s *App) SendMessage(ctx context.Context, req *ServiceUnifiedMessage.SendMessageRequest) (*ServiceUnifiedMessage.SendMessageResponse, error) {

	if s.config.Id == req.ChannelId {
		conn, ok := s.connMap.Get(req.ClientId)
		if ok {
			err := conn.WriteMessage(websocket.TextMessage, req.Message)
			if err != nil {
				return nil, err
			}

		}
	} else if s.pulsar != nil {

		message, _ := json.Marshal(req)

		_, err := s.pulsar.SendMessage(req.ChannelId, message)
		if err != nil {
			return nil, err
		}

	}

	return &ServiceUnifiedMessage.SendMessageResponse{
		Success: true,
	}, nil
}

type App struct {
	ServiceAuth.UnimplementedServiceAuthServer
	ServiceUnifiedMessage.UnifiedMessageServer
	server             *grpc.Server
	config             *AppConfig.AppConfig
	storage            *Storage.Storage
	connMap            *ConnMap.ConnMap
	pulsar             *Pulsar.Client
	stopListenMessages context.CancelFunc
}

func New(storage *Storage.Storage, config *AppConfig.AppConfig, connMap *ConnMap.ConnMap, pulsar *Pulsar.Client) *App {
	server := grpc.NewServer()

	app := &App{
		server:  server,
		config:  config,
		storage: storage,
		connMap: connMap,
		pulsar:  pulsar,
	}

	ServiceAuth.RegisterServiceAuthServer(server, app)
	ServiceUnifiedMessage.RegisterUnifiedMessageServer(server, app)

	if pulsar != nil {

		ctx, cancel := context.WithCancel(context.Background())
		app.stopListenMessages = cancel
		go app.listenPulsarMessage(ctx)

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
	if e.stopListenMessages != nil {
		e.stopListenMessages()
	}
	e.storage.Close()
	e.server.Stop()
}

func (app *App) listenPulsarMessage(ctx context.Context) {

	for msg := range app.pulsar.ListenMessages(ctx, 10) {

		var req ServiceUnifiedMessage.SendMessageRequest
		err := json.Unmarshal(msg, &req)
		if err == nil {
			clientId := req.ClientId

			conn, ok := app.connMap.Get(clientId)
			if ok {
				conn.WriteMessage(websocket.TextMessage, req.Message)
			}

		}

	}
}
