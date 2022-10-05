package server

import "os"

type serverConfig struct {
	// This would be the URL to direct users to after authentication
	appHomepage string

	// This is the unique name given to your app
	appSlug string

	// These are provided by DigitalOcean upon creating your add-on
	appPassword  string
	appSalt      string
	clientSecret string

	// Address this sample server should run on
	serverAddr string
}

func setupServer() *serverConfig {
	config := &serverConfig{
		appSlug:      valueOrDefault("APP_SLUG", "sample_app"),
		appPassword:  valueOrDefault("APP_PASSWORD", ""),
		appSalt:      valueOrDefault("APP_SALT", ""),
		appHomepage:  valueOrDefault("APP_HOMEPAGE", ""),
		clientSecret: valueOrDefault("CLIENT_SECRET", ""),
		serverAddr:   valueOrDefault("SERVER_ADDR", ":8082"),
	}

	return config
}

func valueOrDefault(key string, defaultVal string) string {
	envVar, isSet := os.LookupEnv(key)
	if !isSet {
		return defaultVal
	}
	return envVar
}
