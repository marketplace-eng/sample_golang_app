package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

// DigitalOcean endpoints

// DigitalOcean will send a provisioning request when a user adds
// the add-on to their account
func (s *server) provisionHandler(c echo.Context) error {
	// Parse the request
	s.e.Logger.Info("Got provisioning request")
	ctx := context.Background()
	req := &ProvisioningRequest{}
	err := c.Bind(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, "malformed request: "+err.Error())
	}

	// Create a new account with the given information
	resp, err := s.provisionAccount(ctx, req)
	// If an error occurs, return 422 with message
	if err != nil {
		resp := &ErrorResponse{
			Message: err.Error(),
		}
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	// Trade in the authorization code provided with the provisioning request
	// for a longer-lived access token and permanent refresh token
	err = s.tradeAuthCode(ctx, req.OauthGrant, req.ResourceUUID)
	if err != nil {
		s.e.Logger.Info("Error while trading auth code: " + err.Error())
	}

	// Return a successful response
	return c.JSON(http.StatusOK, resp)
}

// DigitalOcean will send a deprovisioning request when a user removes
// an add-on from their account
func (s *server) deprovisionHandler(c echo.Context) error {
	// Parse the request
	uuid := c.Param("resource_uuid")
	s.e.Logger.Info("Got deprovision request for " + uuid)

	// Deprovision this account
	err := s.deprovisionRequest(context.Background(), uuid)
	if err != nil {
		s.e.Logger.Info("Got " + err.Error())
		_, ok := err.(*NotFoundError)
		// In the event the resource was not found, return a 404
		if ok {
			return c.NoContent(http.StatusNotFound)
		}
		resp := &ErrorResponse{
			Message: err.Error(),
		}
		// If a different error occurs, return a 422 with a message
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	// Return a successful response
	s.e.Logger.Info("Success")
	return c.NoContent(http.StatusOK)
}

// If a user changes their add-on plan, DigitalOcean will send a
// Plan Change request
func (s *server) planChangeHandler(c echo.Context) error {
	// Parse the request
	s.e.Logger.Info("Got change request")
	uuid := c.Param("resource_uuid")

	req := &PlanChangeRequest{}
	err := c.Bind(req)
	if err != nil {
		return c.String(http.StatusBadRequest, "malformed request: "+err.Error())
	}

	// Update the account plan
	err = s.planChange(context.Background(), req, uuid)

	if err != nil {
		_, ok := err.(*NotFoundError)
		// If the account was not found, return a 404
		if ok {
			return c.NoContent(http.StatusNotFound)
		}
		resp := &ErrorResponse{
			Message: err.Error(),
		}
		// If a different error occurs, return a 422
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	// On success, return either 200 with a message or a 204
	return c.NoContent(http.StatusNoContent)
}

// DigitalOcean sends multiple types of notifications when various events
// occur. See documentation for full details.
func (s *server) notificationHandler(c echo.Context) error {
	// Parse the request
	s.e.Logger.Info("Got notification request")
	s.e.Logger.Info(c.Request().Header)

	// Store body data to refill later
	bodyBytes, err := ioutil.ReadAll(c.Request().Body)
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	if err != nil {
		return c.String(http.StatusBadRequest, "malformed request: "+err.Error())
	}

	// Parse request to get Notification type
	var req interface{}
	err = c.Bind(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, "malformed request: "+err.Error())
	}

	// Use type of notification to determine structure of data
	m := req.(map[string]interface{})
	t := m["type"].(string)
	var n Notification
	switch t {
	case Suspended:
		n = &SuspensionNotification{}
	case Reactivated:
		n = &ReactivatedNotification{}
	case DeprovisioningFailed:
		n = &DeprovisioningFailedNotification{}
	case Updated:
		n = &UpdatedNotification{}
	default:
		s.e.Logger.Info("Unknown notification type")
		resp := &ErrorResponse{
			Message: "Unknown notification type",
		}

		// Return a 422 with message if errors occur
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	// Refill request body so we can bind it again
	c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	err = c.Bind(n)
	if err != nil {
		s.e.Logger.Info("Error binding notification: " + err.Error())
		resp := &ErrorResponse{
			Message: err.Error(),
		}

		// Return a 422 with message if errors occur
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	// Pass to the relevant handler
	errs := s.parseNotification(context.Background(), n)

	if len(errs) > 0 {
		resp := &ErrorResponse{
			Message: fmt.Sprintf("Errors occurred: %v", errs),
		}
		// Return a 422 with message if errors occur
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}
	// Return a successful response
	return c.NoContent(http.StatusOK)
}

// When a user accesses the add-on, DigitalOcean sends a single sign-on request
func (s *server) ssoHandler(c echo.Context) error {
	// Parse the request
	s.e.Logger.Info("Got SSO request")

	req := &SsoRequest{}
	err := c.Bind(req)
	if err != nil {
		return c.String(http.StatusBadRequest, "malformed request: "+err.Error())
	}

	// Confirm the given token matches what is expected for this user
	authorized, err := s.authorize(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if !authorized {
		// If it does not, return a 401
		return c.NoContent(http.StatusUnauthorized)
	}

	// Redirect the user to your homepage.
	// Because this example uses a separate front-end, we create
	// a token with the app salt to add as a query parameter. This gets
	// passed to the front-end as part of the redirect, and the front-end will
	// validate it to log the user in.
	token, err := getJWT(s.config.appSalt, req.ResourceUUID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	c.Response().Header().Set("Location", s.config.appHomepage+"?secret="+token)
	return c.NoContent(http.StatusTemporaryRedirect)
}

// Vendor endpoints: for use by this example's front-end

// Called by the front end to verify a given authorization token and
// log a user in. The front-end's half of the SSO request above.
func (s *server) authorizeHandler(c echo.Context) error {
	// Validate the given token
	req := &AuthorizeRequest{}
	err := c.Bind(req)
	if err != nil {
		return c.String(http.StatusBadRequest, "malformed request: "+err.Error())
	}

	authorized, err := validateToken(req.Secret, s.config.appSalt)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// If it is invalid, return a 401
	if !authorized {
		return c.NoContent(http.StatusUnauthorized)
	}

	// Otherwise, return a successful response
	res, err := s.buildAuthResponse(context.Background(), req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, res)
}

// Used to demonstrate sending updated config information to DigitalOcean.
// In this example, a license key is used to represent a vendor's config info.
func (s *server) changeConfig(c echo.Context) error {
	// Parse the request
	s.e.Logger.Info("Got config request")
	uuid := c.QueryParam("uuid")

	// Send the config update to DigitalOcean
	err := s.updateConfig(context.Background(), uuid)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// Return success back to the front end
	return c.NoContent(http.StatusOK)
}
