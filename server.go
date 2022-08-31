package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

const (
	addr = ":8080"
)

type server struct {
	e *echo.Echo
}

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	s := &server{
		e: e,
	}

	// DigitalOcean endpoints

	e.POST("/digitalocean/resources", s.provisionHandler)

	e.DELETE("/digitalocean/resources/:resource_uuid", s.deprovisionHandler)

	e.PUT("/digitalocean/resources/:resource_uuid", s.planChangeHandler)

	e.POST("/digitalocean/notifications", s.notificationHandler)

	e.POST("/digitalocean/sso", s.ssoHandler)

	// Vendor app endpoints (subject to change)

	e.GET("/activities", s.getActivities)

	e.POST("/config", s.changeConfig)

	e.Logger.Fatal(e.Start(addr))
}
