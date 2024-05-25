package jwt

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"auth-microservice/model"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}
type UserClaims struct {
	jwt.StandardClaims
	userEmail string `bson:"email,omitempty"`
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) (*JWTManager, error) {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}, nil
}

// GenerateToken generates a new JWT token and populates it with the user's email.
// username and email are stored in the token as claims.
// userclaims is a struct that contains the claims that will be stored in the token and is used to generate the token.
// The token is signed using the secret key.
func (manager *JWTManager) GenerateToken(user *model.User) (string, error) {
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(manager.tokenDuration).Unix(),
		},
		userEmail: user.Email,
	}

	// creating new token...
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.secretKey))
}

// VerifyToken verifies the JWT token and returns the claims stored in the token.
// The token is verified using the secret key.
//
//	If the token is invalid, an error is returned.
func (manager *JWTManager) VerifyToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(manager.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, err
	}
	return claims, nil
}
