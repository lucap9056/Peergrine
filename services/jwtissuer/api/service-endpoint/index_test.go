package serviceendpoint

import (
	"context"
	ServiceAuth "peergrine/grpc/serviceauth"
	AppConfig "peergrine/jwtissuer/app-config"
	Storage "peergrine/jwtissuer/storage"
	Auth "peergrine/utils/auth"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestVerifyAccessTokenWithServer(t *testing.T) {
	// 設置測試用的 secret 和 token
	secret := []byte("test_secret")
	Iss := "test_issuer"
	UserId := "test_user"
	currentTime := time.Now()
	Iat := currentTime.Unix()
	Exp := currentTime.Add(time.Minute * 1).Unix()

	// 生成測試 token
	token, err := Auth.GenerateBearerToken(Iss, UserId, secret, Iat, Exp)
	if err != nil {
		t.Fatalf("failed to generate bearer token: %v", err)
	}

	storage, err := Storage.New(Iss, "")
	assert.NoError(t, err)
	storage.SaveSecret(secret)

	// 創建 AppConfig 實例
	config := &AppConfig.AppConfig{}

	// 創建測試 server
	server := New(storage, config)

	// 啟動 server 在一個獨立的 goroutine 中
	go func() {
		server.Run("localhost:50051")
	}()

	// 等待 server 啟動
	time.Sleep(time.Second)

	// 創建 gRPC 客戶端連接到測試 server

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := ServiceAuth.NewServiceauthClient(conn)

	// 創建 gRPC 請求
	req := &ServiceAuth.AccessTokenRequest{
		AccessToken: token,
	}

	// 調用 VerifyAccessToken 方法
	res, err := client.VerifyAccessToken(context.Background(), req)
	if err != nil {
		t.Fatalf("failed to verify access token: %v", err)
	}

	// 驗證結果
	assert.Equal(t, Iss, res.Iss, "unexpected issuer")
	assert.Equal(t, Iat, res.Iat, "unexpected issued at time")
	assert.Equal(t, Exp, res.Exp, "unexpected expiration time")
	assert.Equal(t, UserId, res.UserId, "unexpected user id")
}
