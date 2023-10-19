package jwtclaim

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaim struct {
	UserName   string `json:"userName"`
	IsVerified bool   `json:"isVerified"`
	jwt.RegisteredClaims
}

func CreateJwtToken(userName string, isVerified bool) (string, error) {

	claims := UserClaim{
		userName,
		isVerified,
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(240 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	var jwtKey = []byte("my_secret_key")
	jsonWebToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jsonWebToken.SignedString(jwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ExtractVerifyUsername(tokenStr string) (string, bool) {
	hmacSecretString := "my_secret_key"
	hmacSecret := []byte(hmacSecretString)

	token, err := jwt.ParseWithClaims(tokenStr, &UserClaim{}, func(token *jwt.Token) (interface{}, error) {
		return hmacSecret, nil
	}, jwt.WithLeeway(5*time.Second))

	if err != nil {
		return "", false
	}
	if claims, ok := token.Claims.(*UserClaim); ok && token.Valid {
		return claims.UserName, ok
	} else {
		return "", false
	}
}
func ExtractUsername(tokenStr string) (string, bool) {
	hmacSecretString := "my_secret_key"
	hmacSecret := []byte(hmacSecretString)

	token, err := jwt.ParseWithClaims(tokenStr, &UserClaim{}, func(token *jwt.Token) (interface{}, error) {
		return hmacSecret, nil
	}, jwt.WithLeeway(5*time.Second))

	if err != nil {
		return "", false
	}
	if claims, ok := token.Claims.(*UserClaim); ok && token.Valid && claims.IsVerified {
		return claims.UserName, ok
	} else {
		return "", false
	}
}
