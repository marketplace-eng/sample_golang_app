package server

import "os"

type serverConfig struct {
	appSlug      string
	appPassword  string
	appSalt      string
	appHomepage  string
	clientSecret string
}

func setupServer() *serverConfig {
	config := &serverConfig{
		appSlug:      valueOrDefault("APP_SLUG", "sample_app"),
		appPassword:  valueOrDefault("APP_PASSWORD", ""),
		appSalt:      valueOrDefault("APP_SALT", ""),
		appHomepage:  valueOrDefault("APP_HOMEPAGE", ""),
		clientSecret: valueOrDefault("CLIENT_SECRET", ""),
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
