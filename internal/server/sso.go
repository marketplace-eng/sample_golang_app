package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	appSalt     = ""
	appHomepage = ""
)

type SsoRequest struct {
	ResourceUUID string `param:"resource_uuid" form:"resource_uuid"`
	Token        string `param:"token" form:"token"`
	Timestamp    string `param:"timestamp" form:"timestamp"`
	Email        string `param:"user_email" form:"user_email"`
	Id           string `param:"user_id" form:"user_id"`
}

func getJWT() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Minute * 15).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(appSalt)

	return tokenString, err
}

func validateToken(tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return appSalt, nil
	})

	if err != nil {
		return false, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims.VerifyExpiresAt(time.Now().Unix(), true), nil
	} else {
		return false, nil
	}
}

func (s *server) authorize(req *SsoRequest) (bool, error) {
	authorized, err := validToken(req.Token, req.Timestamp, req.ResourceUUID)
	if err != nil {
		return false, err
	} else if !authorized {
		return false, nil
	}

	return true, nil
}

func validToken(token string, timestamp string, uuid string) (bool, error) {
	// has this timestamp expired?
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false, err
	}
	tm := time.Unix(i, 0)
	if time.Since(tm).Minutes() > 2 {
		return false, nil
	}

	// is this token valid?
	decodedToken, err := hex.DecodeString(token)
	if err != nil {
		return false, err
	}
	message := []byte(fmt.Sprintf("%s:%s", timestamp, uuid))

	hash := hmac.New(sha256.New, []byte(appSalt))
	hash.Write(message)

	return hmac.Equal(hash.Sum(nil), []byte(decodedToken)), nil
}
