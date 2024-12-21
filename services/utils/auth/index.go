package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// GenerateBearerToken 生成 Bearer Token。
// 參數:
//
//	iss (string): 令牌的發行者。
//	userId (string): 使用者的唯一標識。
//	secret ([]byte): 用於簽署令牌的密鑰。
//	iat (int64): 令牌的簽發時間（UNIX 時間戳）。
//	exp (int64): 令牌的過期時間（UNIX 時間戳）.
//
// 返回值:
//
//	string: 生成的 Bearer Token。
//	error: 如果生成令牌過程中發生錯誤，則返回錯誤信息。
func GenerateBearerToken(iss string, userId string, channelId int32, secret []byte, iat, exp int64) (string, error) {
	payload := jwt.MapClaims{
		"iss":        iss,
		"iat":        iat,
		"exp":        exp,
		"user_id":    userId,
		"channel_id": channelId,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign bearer token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken 生成 Refresh Token。
// 參數:
//
//	iss (string): 令牌的發行者。
//	userId (string): 使用者的唯一標識。
//	secret ([]byte): 用於簽署令牌的密鑰。
//	iat (time.Time): 令牌的簽發時間。
//
// 返回值:
//
//	string: 生成的 Refresh Token。
//	error: 如果生成令牌過程中發生錯誤，則返回錯誤信息。
func GenerateRefreshToken(iss string, userId string, channelId int32, secret []byte, iat time.Time) (string, error) {
	payload := jwt.MapClaims{
		"iss":        iss,
		"iat":        iat.Unix(),
		"user_id":    userId,
		"channel_id": channelId,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// DecodeToken 解析 token 並返回 claims。
// 參數:
//
//	tokenStr (string): 要解析的 token 字符串。
//	secret ([]byte): 用於驗證 token 的密鑰。
//
// 返回值:
//
//	*jwt.MapClaims: 解析得到的 claims。
//	error: 如果解析 token 過程中發生錯誤，則返回錯誤信息。
func DecodeToken(tokenStr string, secret []byte) (*jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("token is invalid or expired")
	}

	return &claims, nil
}

func Claims2TokenPayload(token string, claims *jwt.MapClaims) (tokenPayload TokenPayload) {

	iat, _ := (*claims)["iat"].(float64)
	exp, _ := (*claims)["exp"].(float64)
	channelId, _ := (*claims)["exp"].(float64)

	return TokenPayload{
		Token:     token,
		Iss:       (*claims)["iss"].(string),
		Iat:       int64(iat),
		Exp:       int64(exp),
		UserId:    (*claims)["user_id"].(string),
		ChannelId: int32(channelId),
	}
}

// ExtractIssuerFromToken 從 token 中提取 iss 欄位。
// 參數:
//
//	tokenString (string): 要解析的 token 字符串。
//
// 返回值:
//
//	string: 提取出的發行者。
//	error: 如果提取過程中發生錯誤，則返回錯誤信息。
func ExtractIssuerFromToken(tokenString string) (string, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}

	payload := parts[1]
	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode payload: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return "", fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if iss, ok := claims["iss"].(string); ok {
		return iss, nil
	}

	return "", fmt.Errorf("iss field not found in token")
}
