package auth_test

import (
	"encoding/base64"
	"encoding/json"
	"peergrine/utils/auth"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

var secret = []byte("secret-key")

func TestGenerateBearerToken(t *testing.T) {
	iss := "test_issuer"
	userId := "user_123"
	iat := time.Now().Unix()
	exp := iat + 3600 // 1 小時後過期

	token, err := auth.GenerateBearerToken(iss, userId, secret, iat, exp)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 驗證生成的 Token
	claims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)
	assert.Equal(t, iss, claims["iss"])
	assert.Equal(t, userId, claims["user_id"])
	assert.Equal(t, iat, int64(claims["iat"].(float64)))
	assert.Equal(t, exp, int64(claims["exp"].(float64)))
}

func TestGenerateRefreshToken(t *testing.T) {
	iss := "test_issuer"
	userId := "user_123"
	iat := time.Now()

	token, err := auth.GenerateRefreshToken(iss, userId, secret, iat)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 驗證生成的 Refresh Token
	claims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)
	assert.Equal(t, iss, claims["iss"])
	assert.Equal(t, userId, claims["user_id"])
	assert.Equal(t, iat.Unix(), int64(claims["iat"].(float64)))
}

func TestDecodeToken(t *testing.T) {
	iss := "test_issuer"
	userId := "user_123"
	iat := time.Now().Unix()
	exp := iat + 3600

	token, err := auth.GenerateBearerToken(iss, userId, secret, iat, exp)
	assert.NoError(t, err)

	claims, err := auth.DecodeToken(token, secret)
	assert.NoError(t, err)
	assert.Equal(t, iss, (*claims)["iss"])
	assert.Equal(t, userId, (*claims)["user_id"])
	assert.Equal(t, iat, int64((*claims)["iat"].(float64)))
	assert.Equal(t, exp, int64((*claims)["exp"].(float64)))

	// 測試錯誤情況
	invalidToken := token + "invalid"
	_, err = auth.DecodeToken(invalidToken, secret)
	assert.Error(t, err)
}

func TestExtractIssuerFromToken(t *testing.T) {
	iss := "test_issuer"
	userId := "user_123"
	iat := time.Now().Unix()
	exp := iat + 3600

	token, err := auth.GenerateBearerToken(iss, userId, secret, iat, exp)
	assert.NoError(t, err)

	extractedIss, err := auth.ExtractIssuerFromToken(token)
	assert.NoError(t, err)
	assert.Equal(t, iss, extractedIss)

	// 測試錯誤情況 - 錯誤格式的 token
	invalidToken := "invalid.token.format"
	_, err = auth.ExtractIssuerFromToken(invalidToken)
	assert.Error(t, err)

	// 測試錯誤情況 - 沒有 iss 欄位
	parts := strings.Split(token, ".")
	payload := parts[1]
	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	assert.NoError(t, err)

	var claims map[string]interface{}
	err = json.Unmarshal(decoded, &claims)
	assert.NoError(t, err)
	delete(claims, "iss") // 刪除 iss 欄位

	// 重新編碼 payload
	updatedPayload, err := json.Marshal(claims)
	assert.NoError(t, err)
	parts[1] = base64.RawURLEncoding.EncodeToString(updatedPayload)
	modifiedToken := strings.Join(parts, ".")

	_, err = auth.ExtractIssuerFromToken(modifiedToken)
	assert.Error(t, err)
}
