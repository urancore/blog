package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenLivetime = 12 * time.Hour // TODO: add in config
	tokenKey      = "z!mZ:S~r7a$nac];wX_t*+I9?ZA;`9[vS!}S5#9Zp|e~qjp%jqUA%3'<W.6k82J"
)

type tokenClaims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"user_id"`
}

func GenerateToken(userID int64, username string, email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenLivetime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	})
	return token.SignedString([]byte(tokenKey))
}

func ParseToken(accessToken string) (int64, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(tokenKey), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, errors.New("invalid token claims type")
	}
	return claims.UserID, nil
}
