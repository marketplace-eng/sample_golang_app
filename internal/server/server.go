package server

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

const (
	addr = ":8082"
)

type server struct {
	e  *echo.Echo
	db *pgxpool.Pool
}

func StartServer(ctx context.Context, db *pgxpool.Pool) {
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	s := &server{
		e:  e,
		db: db,
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
