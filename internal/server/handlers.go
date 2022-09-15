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
	req := &ProvisioningRequest{}
	err := c.Bind(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, "malformed request: "+err.Error())
	}

	resp, err := s.provisionAccount(context.Background(), req)
	// If an error occurs, return 422 with message
	if err != nil {
		resp := &ErrorResponse{
			Message: err.Error(),
		}
		return c.JSON(http.StatusUnprocessableEntity, resp)
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
	return c.String(http.StatusOK, fmt.Sprintf("I'll change %s\n", uuid))
}

func (s *server) notificationHandler(c echo.Context) error {
	s.e.Logger.Info("Got notification request")
	return c.String(http.StatusOK, "I got a notification!\n")
}

func (s *server) ssoHandler(c echo.Context) error {
	s.e.Logger.Info("Got sso request")
	return c.String(http.StatusOK, "This will eventually do the complicated SSO dance\n")
}

// Vendor app endpoints (subject to change)

func (s *server) getActivities(c echo.Context) error {
	s.e.Logger.Info("Got activities request")
	return c.String(http.StatusOK, "I'll fetch your activities\n")
}

func (s *server) changeConfig(c echo.Context) error {
	s.e.Logger.Info("Got config request")
	return c.String(http.StatusOK, "I'll send config data back to DO\n")
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
