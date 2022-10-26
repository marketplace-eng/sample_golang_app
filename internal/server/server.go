package server

import (
	"context"
	"crypto/subtle"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type server struct {
	e      *echo.Echo
	db     *pgxpool.Pool
	config *serverConfig
}

// Start the server for our example application.
func StartServer(ctx context.Context, db *pgxpool.Pool) {
	e := echo.New()

	config := setupServer()

	// DigitalOcean will call your app with basic auth headers, using slug and password set up on app creation.
	e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		// Uses constant time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(username), []byte(config.appSlug)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(config.appPassword)) == 1 {
			return true, nil
		}
		return false, nil
	}))
	e.Logger.SetLevel(log.INFO)

	s := &server{
		e:      e,
		db:     db,
		config: config,
	}

	// DigitalOcean endpoints

	e.GET("/", s.ssoHandler)

	e.POST("/digitalocean/resources", s.provisionHandler)

	e.DELETE("/digitalocean/resources/:resource_uuid", s.deprovisionHandler)

	e.PUT("/digitalocean/resources/:resource_uuid", s.planChangeHandler)

	e.POST("/digitalocean/notifications", s.notificationHandler)

	e.POST("/digitalocean/sso", s.ssoHandler)

	// Vendor endpoints: for use by this example's front-end

	e.POST("/config/:uuid", s.changeConfig)

	e.POST("/authorize/sso", s.authorizeHandler)

	e.Logger.Fatal(e.Start(config.serverAddr))
}
