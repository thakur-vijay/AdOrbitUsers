package config

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	Port         string
	GRPCPort     string
	DatabaseURL  string
	JWTPublicKey string
	RabbitMQURL  string
}

func Load() (*Config, error) {
	port := normalizePort(getEnv("APP_PORT", "8080"))
	grpcPort := normalizePort(getEnv("GRPC_PORT", "9090"))

	cfg := &Config{
		Port:         port,
		GRPCPort:     grpcPort,
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		JWTPublicKey: strings.ReplaceAll(os.Getenv("JWT_PUBLIC_KEY"), `\n`, "\n"),
		RabbitMQURL:  getEnv("RABBITMQ_URL", ""),
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is not set")
	}

	if cfg.JWTPublicKey == "" {
		return nil, errors.New("JWT_PUBLIC_KEY is not set")
	}

	if cfg.RabbitMQURL == "" {
		return nil, errors.New("RABBITMQ_URL is not set")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func normalizePort(port string) string {
	if port == "" {
		return ":8080"
	}

	if strings.Contains(port, ":") {
		return port
	}

	return ":" + port
}
