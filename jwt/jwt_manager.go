package jwt

import (
	"time"

	"auth-microservice/model"

	"github.com/dgrijalva/jwt-go"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}
type UserClaims struct {
	jwt.StandardClaims
	UserEmail string
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) (*JWTManager, error) {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}, nil
}
func (manager *JWTManager) GenerateToken(user *model.User) (string, error) {
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(manager.tokenDuration).Unix(),
		},
		UserEmail: user.Email,
	}
	// creating new token...
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.secretKey))
}
