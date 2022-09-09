package server

import "sample_app/models"

type ProvisioningRequest struct {
	// User selected app slug as provided by the vendor during vendor registration
	AppSlug string `json:"app_slug"`

	// User selected plan slug as provided by the vendor during vendor registration
	PlanSlug string `json:"plan_slug"`

	// DigitalOcean generated UUID for identifying this specific resource
	ResourceUUID string `json:"uuid"`

	// Customizable metadata that a DigitalOcean user can set for this specific resource
	Metadata ProvisioningMetadata `json:"metadata"`

	// An obfuscated email pointing to the user’s email address. Anything sent to this email will be
	// forwarded to the user.
	Email string `json:"email"`

	// DigitalOcean obfuscated ID that will uniquely identify the user's team. This is useful to know
	// when the same DigitalOcean team provisions multiple resources for your Add-On.
	TeamID string `json:"creator_id"`

	// Your app can asynchronously exchange this authorization_code for an access_token and
	// refresh_token upon success. With the access_token, you can modify this specific resource within
	// DigitalOcean.
	OauthGrant OauthGrant `json:"oauth_grant"`
}

type ProvisioningMetadata struct {
	Language        string `json:"language"`
	EmailPreference bool   `json:"email_preference"`
}

type OauthGrant struct {
	CodeType   string `json:"type"`
	Code       string `json:"code"`
	Expires_at int    `json:"expires_at"`
}

type ProvisioningResponse struct {
	// Required: An immutable value for DigitalOcean to reference this resource within your app.
	// This example uses the resource’s UUID from the originating request.
	Id string `json:"id"`

	// The variables necessary to enable the DigitalOcean user to utilize your app
	// (endpoints, credentials, etc). They will be displayed to the user from the Add-On pages
	// on DigitalOcean. The variables will be prefixed with your Add-On’s Configuration Variable
	// Prefix config_vars_prefix.
	Config ProvisioningConfig `json:"config"`

	// Optional for success, recommended for failure
	Message string `json:"message"`
}

type ProvisioningConfig struct {
	LicenseKey string `json:"LICENSE_KEY"`
}

func provisionAccount(req *ProvisioningRequest) (*ProvisioningResponse, error) {
	acc := createAccount(req)

	// database.create(acc)

	resp := &ProvisioningResponse{
		Id: acc.ResourceUUID,
		Config: ProvisioningConfig{
			LicenseKey: acc.LicenseKey,
		},
		Message: "Account provisioning succeeded!",
	}
	return resp, nil
}

func createAccount(req *ProvisioningRequest) *models.Account {
	acc := &models.Account{
		Name:            req.TeamID,
		Email:           req.Email,
		AppSlug:         req.AppSlug,
		PlanSlug:        req.PlanSlug,
		ResourceUUID:    req.ResourceUUID,
		Language:        req.Metadata.Language,
		EmailPreference: req.Metadata.EmailPreference,
		Source:          "DigitalOcean",
		Status:          models.Active,
		LicenseKey:      "",
	}
	return acc
}
