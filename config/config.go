package config

import "os"

const (
	DefaultJWTSecret = "my_super_secret_jwt_key"
)

func GetJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return DefaultJWTSecret
	}
	return secret
}