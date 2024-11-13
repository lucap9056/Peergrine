package keys

func RefreshToken(refreshToken string) string {
	return "refresh_token:" + refreshToken
}

func Secret(secretId string) string {
	return "secret:" + secretId
}
