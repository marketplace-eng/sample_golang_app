package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

type dbConfig struct {
	dbUsername string
	dbPassword string
	dbHost     string
	dbPort     string
	dbName     string
}

func OpenDB() (*pgxpool.Pool, error) {
	config := setupDB()
	dataSourceName := "postgresql://" + config.dbUsername + ":" + config.dbPassword + "@" + config.dbHost + ":" + config.dbPort + "/" + config.dbName
	conn, err := pgxpool.Connect(context.Background(), dataSourceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return nil, err
	}

	return conn, nil
}

func setupDB() *dbConfig {
	config := &dbConfig{
		dbUsername: valueOrDefault("DB_USERNAME", "postgres"),
		dbPassword: valueOrDefault("DB_PASSWORD", "example"),
		dbHost:     valueOrDefault("DB_HOST", "localhost"),
		dbPort:     valueOrDefault("DB_PORT", "5431"),
		dbName:     valueOrDefault("DB_NAME", "postgres"),
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
