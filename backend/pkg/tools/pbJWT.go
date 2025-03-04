package tools

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// PocketBaseJWTPayload represents the payload of a JWT token used by PocketBase.
type PocketBaseJWTPayload struct {
	CollectionId string `json:"collectionId"`
	Id           string `json:"id"`
	Exp          int    `json:"exp"`
	Type         string `json:"type"`
	Refreshable  bool   `json:"refreshable"`
	jwt.RegisteredClaims
}

// GetUserIdFromJWT extracts the user ID from a JWT token.
//
// It returns an empty string and an error if the token is invalid or does not contain a user ID.
func GetUserIdFromJWT(tokenString string) (string, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &PocketBaseJWTPayload{})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*PocketBaseJWTPayload); ok {
		return claims.Id, nil
	}

	return "", errors.New("invalid token or claims")
}

// VerifyJWTExpiration checks if the JWT token is still valid based on its expiration time.
//
// It returns true if the token is still valid, and an error if the token is invalid.
func VerifyJWTExpiration(tokenString string) (bool, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &PocketBaseJWTPayload{})
	if err != nil {
		return false, err
	}

	if claims, ok := token.Claims.(*PocketBaseJWTPayload); ok {
		return time.Now().Unix() < int64(claims.Exp), nil
	}
	return false, errors.New("invalid token or claims")
}
