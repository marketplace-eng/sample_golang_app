package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// DigitalOcean endpoints

func (s *server) provisionHandler(c echo.Context) error {
	s.e.Logger.Info("Got provision request")
	return c.String(http.StatusOK, "I'll provision something\n")
}

func (s *server) deprovisionHandler(c echo.Context) error {
	s.e.Logger.Info("Got deprovision request")
	uuid := c.Param("resource_uuid")
	return c.String(http.StatusOK, fmt.Sprintf("I'll deprovision %s\n", uuid))
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
