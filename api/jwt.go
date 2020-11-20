package api

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

var SigningSecret = "xjh7VdjjGgC38XzDvHquQdc3Z5Gs2CNdW6kk3rqPuUNp2vnRG3rXbv33mKbZxHAE"

type UserJWTClaims struct {
	jwt.StandardClaims
	UserOpen
}

func createUserJWTToken(user UserOpen, secret string, timeout time.Duration) (string, error) {
	claims := UserJWTClaims{UserOpen: user}
	claims.IssuedAt = time.Now().Unix()
	claims.ExpiresAt = time.Now().Add(timeout).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func readUserJWTToken(tokenStr, secret string) (*UserJWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("bad signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*UserJWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("bad claims type or token is not valid")
}
