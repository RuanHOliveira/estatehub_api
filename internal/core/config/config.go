package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	AppConfig  *appConfig
	PgConfig   *PgConfig
	AuthConfig *AuthConfig
}

type appConfig struct {
	AppEnv    string
	AppPort   int
}

type PgConfig struct {
	PgHost string
	PgPort int
	PgUser string
	PgPass string
	PgName string
}

type AuthConfig struct {
	JwtSecret string
}

func Load() *Config {
	cfg := &Config{
		AppConfig: &appConfig{
			AppEnv:    getEnv("APP_ENV", "dev"),
			AppPort:   getIntEnvOrPanic("APP_PORT"),
		},
		PgConfig: &PgConfig{
			PgHost: getStringEnvOrPanic("PG_HOST"),
			PgPort: getIntEnvOrPanic("PG_PORT"),
			PgUser: getStringEnvOrPanic("PG_USER"),
			PgPass: getStringEnvOrPanic("PG_PASS"),
			PgName: getStringEnvOrPanic("PG_NAME"),
		},
		AuthConfig: &AuthConfig{
			JwtSecret: getStringEnvOrPanic("JWT_SECRET"),
		},
	}

	if len(cfg.AuthConfig.JwtSecret) < 32 {
		log.Fatal("FATAL: JWT_SECRET must be at least 32 characters long")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getStringEnvOrPanic(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("FATAL: Missing required environment variable %s", key)
	}
	return v
}

func getIntEnvOrPanic(key string) int {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("FATAL: Missing required environment variable %s", key)
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("FATAL: Invalid integer value for %s: %v", key, err)
	}
	return i
}
