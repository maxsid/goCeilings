package api

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/maxsid/goCeilings/server/common"
)

var SigningSecret = "xjh7VdjjGgC38XzDvHquQdc3Z5Gs2CNdW6kk3rqPuUNp2vnRG3rXbv33mKbZxHAE"

type userJWTClaims struct {
	jwt.StandardClaims
	common.UserBasic
}

func createUserJWTToken(user common.UserBasic, secret string, timeout time.Duration) (string, error) {
	claims := userJWTClaims{UserBasic: user}
	claims.IssuedAt = time.Now().Unix()
	claims.ExpiresAt = time.Now().Add(timeout).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func readUserJWTToken(tokenStr, secret string) (*userJWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &userJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("bad signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*userJWTClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("bad claims type or token is not valid")
}
