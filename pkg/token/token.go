package token

import (
	"fmt"
	"full_backend_practice/pkg/database"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenClaim struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var (
	AccessSecret  = []byte("access_secret_example_change_me")
	RefreshSecret = []byte("refresh_secret_example_change_me")
)

func TokenCreate(user *database.User) (accessToken string, refreshToken string, err error) {
	now := time.Now()
	accessClaim := TokenClaim{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	}
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaim).SignedString(AccessSecret)
	if err != nil {
		return "", "", err
	}
	refreshClaim := TokenClaim{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		},
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaim).SignedString(RefreshSecret)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil

}
func ParseToken(tokenString string, secret []byte) (*TokenClaim, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaim{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*TokenClaim); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
