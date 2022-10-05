package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
)

const (
	UpdateLicenseKeySQL = `
	UPDATE accounts
	SET license_key=$2
	WHERE resource_uuid=$1;
	`
)

func (s *server) updateConfig(ctx context.Context, uuid string) error {
	configURL := "https://api.digitalocean.com:443/v2/add-ons/resources/" + uuid + "/config"

	token, err := s.getAccessToken(ctx, uuid)
	if err != nil {
		return err
	}

	licenseKey := newLicenseKey()

	//update license key
	err = s.updateLicenseKey(ctx, licenseKey, uuid)
	if err != nil {
		return err
	}

	configReq := ConfigUpdate{
		Config: ProvisioningConfig{
			LicenseKey: licenseKey,
		},
	}

	jsonBody, err := json.Marshal(configReq)
	if err != nil {
		s.e.Logger.Info("error converting config update request to json: " + err.Error())
		return err
	}
	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPatch, configURL, bodyReader)
	if err != nil {
		s.e.Logger.Info("error creating config HTTP request: " + err.Error())
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		s.e.Logger.Info("error making config http request: " + err.Error())
		return err
	}

	if res.StatusCode >= 400 {
		s.e.Logger.Info("got bad status response from config update")
		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			s.e.Logger.Info("could not read error response: " + err.Error())
			return err
		}
		s.e.Logger.Info("response body: " + string(resBody))
	}

	return nil
}

func (s *server) updateLicenseKey(ctx context.Context, licenseKey string, uuid string) error {
	_, err := s.db.Exec(ctx, UpdateLicenseKeySQL,
		uuid,
		licenseKey,
	)

	if err != nil {
		s.e.Logger.Info("error updating license key: " + err.Error())
		return err
	}

	return nil
}

func newLicenseKey() string {
	return uuid.New().String()
}
