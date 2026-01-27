package config

import "os"

type Config struct {
	DatabaseURL string
	GRPCPort    string
	HTTPPort    string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/dbname?sslmode=disable"),
		GRPCPort:    getEnv("GRPC_PORT", "9090"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
