package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Token struct {
	// Used to access the DigitalOcean API scoped to a single resource. Normally expires
	// every 8 hours, but may expire early in certain circumstances.
	AccessToken string `json:"access_token"`

	// Valid for the lifetime of the resource and can be exchanged for a new access_token
	// as many times as needed using a valid OAuth client_secret.
	RefreshToken string `json:"refresh_token"`

	// The number of seconds the access_token is valid for. The refresh_token is used to
	// acquire a new access_token.
	ExpiresIn int64 `json:"expires_in"`

	// Time (in seconds since the epoch) this token will expire
	ExpiresAt int64

	// The token type is used in the Authorization header of requests to the DigitalOcean API
	TokenType string `json:"token_type"`
}

type AuthCodeRequest struct {
	// The authorization code provided during the provisioning request
	Code string `json:"code"`

	// Type of code
	GrantType string `json:"grant_type"`

	// The preshared secret that is associated with your Add-On
	Secret string `json:"client_secret"`
}

type RefreshRequest struct {
	// Type of code
	GrantType string `json:"grant_type"`

	// The authorization code provided during the provisioning request
	RefreshToken string `json:"refresh_token"`

	// The preshared secret that is associated with your Add-On
	ClientSecret string `json:"client_secret"`
}

const (
	digitaloceanTokenAPI = "https://api.digitalocean.com:443/v2/add-ons/oauth/token"

	InsertTokenSQL = `
	INSERT INTO tokens (resource_uuid, access_token, refresh_token, expires_at)
	VALUES ($1, $2, $3, $4, $5);
	`
	GetTokenSQL = `
	SELECT * FROM tokens WHERE resource_uuid=$1;
	`

	UpdateTokenSQL = `
	UPDATE tokens
	SET access_token=$2, refresh_token=$3, expires_at=$4
	WHERE id=$1;
	`
)

func (s *server) tradeAuthCode(ctx context.Context, oauth OauthGrant, uuid string) error {
	exchangeReq := AuthCodeRequest{
		Code:      oauth.Code,
		GrantType: "authorization_code",
		Secret:    clientSecret,
	}

	jsonBody, err := json.Marshal(exchangeReq)
	if err != nil {
		return err
	}

	token, err := makeTokenRequest(jsonBody)
	if err != nil {
		return err
	}

	err = s.saveToken(ctx, token, uuid)
	if err != nil {
		return err
	}

	return nil
}

func (s *server) saveToken(ctx context.Context, token *Token, uuid string) error {
	_, err := s.db.Exec(ctx, InsertTokenSQL,
		uuid,
		token.AccessToken,
		token.RefreshToken,
		time.Now().Unix()+token.ExpiresIn,
	)

	if err != nil {
		s.e.Logger.Error("Unable to save tokens: " + err.Error())
		return err
	}
	return nil
}

func (s *server) getAccessToken(ctx context.Context, uuid string) (string, error) {
	// get access token, refresh token, expire time from db for this uuid
	token, err := s.readTokens(ctx, uuid)
	if err != nil {
		return "", err
	}

	// if token is expired, send refresh token in
	if token.ExpiresAt < time.Now().Unix() {
		token, err = s.refreshToken(ctx, token, uuid)
		if err != nil {
			return "", err
		}
	}

	// return access token
	return token.AccessToken, nil
}

func (s *server) readTokens(ctx context.Context, uuid string) (*Token, error) {
	token := &Token{}
	err := s.db.QueryRow(ctx, GetTokenSQL, uuid).Scan(token)
	if err != nil {
		s.e.Logger.Error("Unable to fetch tokens: " + err.Error())
		return nil, err
	}
	return token, nil
}

func (s *server) refreshToken(ctx context.Context, token *Token, uuid string) (*Token, error) {
	refreshReq := RefreshRequest{
		GrantType:    "refresh_token",
		RefreshToken: token.RefreshToken,
		ClientSecret: clientSecret,
	}

	jsonBody, err := json.Marshal(refreshReq)
	if err != nil {
		return nil, err
	}

	token, err = makeTokenRequest(jsonBody)
	if err != nil {
		return nil, err
	}

	err = s.saveToken(ctx, token, uuid)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func makeTokenRequest(jsonBody []byte) (*Token, error) {
	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, digitaloceanTokenAPI, bodyReader)
	if err != nil {
		fmt.Printf("error creating HTTP request: %s\n", err)
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return nil, err
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		return nil, err
	}

	resp := &Token{}
	err = json.Unmarshal(resBody, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
