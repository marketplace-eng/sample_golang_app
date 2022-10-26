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

	e.Logger.SetLevel(log.INFO)

	s := &server{
		e:      e,
		db:     db,
		config: config,
	}

	// DigitalOcean will call your app with basic auth headers, using slug and password set up on app creation.
	authedEndpoints := e.Group("/")
	authedEndpoints.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		// Uses constant time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(username), []byte(config.appSlug)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(config.appPassword)) == 1 {
			return true, nil
		}
		return false, nil
	}))

	// DigitalOcean endpoints

	e.GET("/", s.ssoHandler)

	authedEndpoints.POST("/digitalocean/resources", s.provisionHandler)

	authedEndpoints.DELETE("/digitalocean/resources/:resource_uuid", s.deprovisionHandler)

	authedEndpoints.PUT("/digitalocean/resources/:resource_uuid", s.planChangeHandler)

	authedEndpoints.POST("/digitalocean/notifications", s.notificationHandler)

	authedEndpoints.POST("/digitalocean/sso", s.ssoHandler)

	// Vendor endpoints: for use by this example's front-end

	authedEndpoints.POST("/config/:uuid", s.changeConfig)

	authedEndpoints.POST("/authorize/sso", s.authorizeHandler)

	e.Logger.Fatal(e.Start(config.serverAddr))
}
