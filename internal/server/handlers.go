package server

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

// DigitalOcean endpoints

func (s *server) provisionHandler(c echo.Context) error {
	s.e.Logger.Info("Got provisioning request")
	ctx := context.Background()
	req := &ProvisioningRequest{}
	err := c.Bind(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, "malformed request: "+err.Error())
	}

	resp, err := s.provisionAccount(ctx, req)
	// If an error occurs, return 422 with message
	if err != nil {
		resp := &ErrorResponse{
			Message: err.Error(),
		}
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	err = s.tradeAuthCode(ctx, req.OauthGrant, req.ResourceUUID)
	if err != nil {
		s.e.Logger.Info("Error while trading auth code: " + err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *server) deprovisionHandler(c echo.Context) error {
	uuid := c.Param("resource_uuid")
	s.e.Logger.Info("Got deprovision request for " + uuid)
	err := s.deprovisionRequest(context.Background(), uuid)
	if err != nil {
		s.e.Logger.Info("Got " + err.Error())
		_, ok := err.(*NotFoundError)
		if ok {
			return c.NoContent(http.StatusNotFound)
		}
		resp := &ErrorResponse{
			Message: err.Error(),
		}
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}
	s.e.Logger.Info("Success")
	return c.NoContent(http.StatusOK)
}

func (s *server) planChangeHandler(c echo.Context) error {
	s.e.Logger.Info("Got change request")
	uuid := c.Param("resource_uuid")

	req := &PlanChangeRequest{}
	err := c.Bind(req)
	if err != nil {
		return c.String(http.StatusBadRequest, "malformed request: "+err.Error())
	}

	err = s.planChange(context.Background(), req, uuid)

	if err != nil {
		_, ok := err.(*NotFoundError)
		if ok {
			return c.NoContent(http.StatusNotFound)
		}
		resp := &ErrorResponse{
			Message: err.Error(),
		}
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *server) notificationHandler(c echo.Context) error {
	s.e.Logger.Info("Got notification request")
	req := &Notification{}
	err := c.Bind(req)
	if err != nil {
		return c.String(http.StatusBadRequest, "malformed request: "+err.Error())
	}

	errs := s.parseNotification(context.Background(), req)
	if len(errs) > 0 {
		resp := &ErrorResponse{
			Message: fmt.Sprintf("Errors occurred: %v", errs),
		}
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}
	return c.NoContent(http.StatusOK)
}

func (s *server) ssoHandler(c echo.Context) error {
	s.e.Logger.Info("Got SSO request")

	req := &SsoRequest{}
	err := c.Bind(req)
	if err != nil {
		return c.String(http.StatusBadRequest, "malformed request: "+err.Error())
	}

	authorized, err := s.authorize(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if !authorized {
		return c.NoContent(http.StatusUnauthorized)
	}

	token, err := getJWT(s.config.appSalt)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	c.Response().Header().Set("Location", s.config.appHomepage+"?secret="+token)
	return c.NoContent(http.StatusTemporaryRedirect)
}

// Vendor app endpoints (subject to change)

func (s *server) authorizeHandler(c echo.Context) error {
	token := c.QueryParam("secret")
	authorized, err := validateToken(token, s.config.appSalt)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	if !authorized {
		return c.NoContent(http.StatusUnauthorized)
	}

	return c.NoContent(http.StatusOK)
}

func (s *server) changeConfig(c echo.Context) error {
	s.e.Logger.Info("Got config request")
	uuid := c.QueryParam("uuid")
	err := s.updateConfig(context.Background(), uuid)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

// for debugging requests
func (s *server) logBody(req http.Request) string {
	defer req.Body.Close()

	b, err := io.ReadAll(req.Body)
	if err != nil {
		s.e.Logger.Fatal(err)
	}

	return string(b)
}
