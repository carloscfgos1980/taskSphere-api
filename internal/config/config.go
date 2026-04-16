package config

import (
	"errors"
	"os"

	"github.com/carloscfgos1980/taskSphere-api/internal/database"

	"github.com/joho/godotenv"
)

var (
	ErrMissingDatabaseURL = errors.New("missing database URL")
	ErrMissingPort        = errors.New("missing port")
	ErrMissingJWT         = errors.New("missing JWT secret")
)

type Config struct {
	DB        *database.Queries
	DB_URL    string
	Port      string
	JWTSecret string
}

func LoadConfig() (*Config, error) {
	// Try common .env locations (project root and cmd/ execution path).
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load("../.env.local")

	DB_URL := os.Getenv("DB_URL")
	if DB_URL == "" {
		DB_URL = os.Getenv("DATABASE_URL")
	}
	if DB_URL == "" {
		return nil, ErrMissingDatabaseURL
	}

	Port := os.Getenv("PORT")
	if Port == "" {
		return nil, ErrMissingPort
	}

	JWTSecret := os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		return nil, ErrMissingJWT
	}

	// Return the configuration struct with the loaded values
	return &Config{
		DB_URL:    DB_URL,
		Port:      Port,
		JWTSecret: JWTSecret,
	}, nil
}
