package server

/**
 * Everything in this file is for interacting with our sample front-end.
 * It will likely not be useful for your app or service.
 */

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

/**
 * This is what our sample front-end will send to this app when someone tries signing in
 */
type AuthorizeRequest struct {
	Secret string `json:"secret"`
}

/**
 * This is what our smaple front-end will expect to get back from an authorize request
 */
type AuthorizeResponse struct {
	AccessToken  string    `json:"access_token"`
	Email        string    `json:"email"`
	AppSlug      string    `json:"app_slug"`
	PlanSlug     string    `json:"plan_slug"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
	ResourceUUID string    `json:"resource_uuid"`
	Message      string    `json:"message"`
}

const (
	GetAccountDataSQL = `
	SELECT email, app_slug, plan_slug, created_at, modified_at FROM accounts WHERE resource_uuid=$1;
	`
)

// Create and sign a JWT with a secret salt to give front-end in order to
// verify authorization of a user
func getJWT(salt string, uuid string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Minute * 15).Unix(),
		"uuid": uuid,
	})

	// Sign and get the complete encoded token as a string using the salt
	tokenString, err := token.SignedString(salt)

	return tokenString, err
}

// Check that a given JWT is still valid and untampered with
func validateToken(tokenString string, salt string) (bool, error) {
	// Parse the given token to ensure it is signed correctly and unmodified
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return salt, nil
	})

	if err != nil {
		return false, err
	}

	// Verify token is still valid and unexpired
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims.VerifyExpiresAt(time.Now().Unix(), true), nil
	} else {
		return false, nil
	}
}

// Construct a response to an auth request from the front-end
func (s *server) buildAuthResponse(ctx context.Context, req *AuthorizeRequest) ([]byte, error) {
	claims, err := getClaims(req.Secret, s.config.appSalt)

	if err != nil {
		return nil, err
	}

	uuid := string(claims["uuid"].(string))
	accessToken, err := s.getAccessToken(ctx, uuid)
	if err != nil {
		return nil, err
	}

	resp := &AuthorizeResponse{}
	err = s.db.QueryRow(ctx, GetAccountDataSQL, uuid).Scan(resp)
	if err != nil {
		return nil, err
	}

	resp.AccessToken = accessToken
	resp.Message = "Welcome to your dashboard!"
	resp.ResourceUUID = uuid

	respJson, err := json.Marshal(resp)
	return respJson, err
}

func getClaims(tokenString string, salt string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return salt, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("invalid JWT")
	}
}
