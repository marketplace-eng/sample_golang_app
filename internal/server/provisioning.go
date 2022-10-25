package server

import (
	"context"
	"sample_app/models"

	"github.com/jackc/pgx/v4"
)

type ProvisioningRequest struct {
	// A name for the resource that the user provided during the provisioning flow.
	Name string `json:"resource_name"`

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

type ConfigUpdate struct {
	Config ProvisioningConfig `json:"config"`
}

const (
	GetAccountSQL = `
	SELECT id FROM accounts WHERE resource_uuid=$1;
	`

	InsertAccountSQL = `
	INSERT INTO accounts (name, email, app_slug, plan_slug, resource_uuid, language, email_preference, source, status, license_key) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
	`

	UpdateAccountSQL = `
	UPDATE accounts
	SET name=$2, email=$3, app_slug=$4, plan_slug=$5, language=$6, email_preference=$7, status=$8, license_key=$9
	WHERE id=$1;
	`
)

// When a user adds your add-on to their account, DigitalOcean will send you a
// provisioning request with user information for you to create an account in your application
func (s *server) provisionAccount(ctx context.Context, req *ProvisioningRequest) (*ProvisioningResponse, error) {
	licenseKey := newLicenseKey()
	var id int

	// Check if this account UUID has previously provisioned an account
	err := s.db.QueryRow(ctx, GetAccountSQL, req.ResourceUUID).Scan(&id)
	if err == pgx.ErrNoRows {
		// If not, create a new account for them
		_, err = s.db.Exec(ctx, InsertAccountSQL,
			req.Name,
			req.Email,
			req.AppSlug,
			req.PlanSlug,
			req.ResourceUUID,
			req.Metadata.Language,
			req.Metadata.EmailPreference,
			"DigitalOcean",
			models.Active,
			licenseKey,
		)
	} else if err == nil {
		// If so, update the existing account
		_, err = s.db.Exec(ctx, UpdateAccountSQL,
			id,
			req.Name,
			req.Email,
			req.AppSlug,
			req.PlanSlug,
			req.Metadata.Language,
			req.Metadata.EmailPreference,
			models.Active,
			licenseKey,
		)
	} else {
		s.e.Logger.Error("Unable to query for account presence: " + err.Error())
		return nil, err
	}

	if err != nil {
		s.e.Logger.Error("Unable to provision account: " + err.Error())
		return nil, err
	}

	// Any user config information should be contained in the provisioning response.
	// Our example uses license keys as sample config information.
	resp := &ProvisioningResponse{
		Id: req.ResourceUUID,
		Config: ProvisioningConfig{
			LicenseKey: licenseKey,
		},
		Message: "Account provisioning succeeded!",
	}
	return resp, nil
}
