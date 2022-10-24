package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

/**
 * This is what DigitalOcean will send to this app when someone tries signing in
 */
type SsoRequest struct {
	ResourceUUID string `param:"resource_uuid" form:"resource_uuid"`
	Token        string `param:"token" form:"token"`
	Timestamp    string `param:"timestamp" form:"timestamp"`
	Email        string `param:"user_email" form:"user_email"`
	Id           string `param:"user_id" form:"user_id"`
}

// Validate a token included in a DigitalOcean SSO Request
func (s *server) authorize(req *SsoRequest) (bool, error) {
	authorized, err := validToken(req.Token, req.Timestamp, req.ResourceUUID, s.config.appSalt)
	if err != nil {
		return false, err
	} else if !authorized {
		return false, nil
	}

	return true, nil
}

// Check that a given token matches the expected timestamp and app salt in order to
// determine its validity
func validToken(token string, timestamp string, uuid string, salt string) (bool, error) {
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

	hash := hmac.New(sha256.New, []byte(salt))
	hash.Write(message)

	return hmac.Equal(hash.Sum(nil), []byte(decodedToken)), nil
}
