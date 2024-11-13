package serviceendpoint

import (
	"context"
	"net"
	ServiceAuth "peergrine/grpc/serviceauth"
	AppConfig "peergrine/jwtissuer/app-config"
	Storage "peergrine/jwtissuer/storage"
	Auth "peergrine/utils/auth"

	"google.golang.org/grpc"
)

type AuthServiceServer struct {
	storage *Storage.Storage
	ServiceAuth.UnimplementedServiceauthServer
}

func (s *AuthServiceServer) VerifyAccessToken(ctx context.Context, req *ServiceAuth.AccessTokenRequest) (*ServiceAuth.TokenResponse, error) {

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
		Iss:    (*claims)["iss"].(string),
		Iat:    int64(iat),
		Exp:    int64(exp),
		UserId: (*claims)["user_id"].(string),
	}

	return &res, nil
}

type Server struct {
	*grpc.Server
	service *AuthServiceServer
}

func New(storage *Storage.Storage, config *AppConfig.AppConfig) *Server {
	server := grpc.NewServer()
	service := &AuthServiceServer{
		storage: storage,
	}

	ServiceAuth.RegisterServiceauthServer(server, service)

	return &Server{
		server,
		service,
	}
}

func (e *Server) Run(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return e.Serve(listener)
}

func (e *Server) Close() {
	e.service.storage.Close()
	e.Server.Stop()
}
