package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Address        string
	NodeAPIURL     string
	JWTSecret      string
	JWTIssuer      string
	MaxMatrixDim   int
	RequestTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Address:        env("HTTP_ADDR", ":8080"),
		NodeAPIURL:     env("NODE_API_URL", "http://localhost:3000"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		JWTIssuer:      env("JWT_ISSUER", "reto-tecnico"),
		MaxMatrixDim:   200,
		RequestTimeout: 5 * time.Second,
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET es obligatorio")
	}

	if raw := os.Getenv("MAX_MATRIX_DIM"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return Config{}, fmt.Errorf("MAX_MATRIX_DIM debe ser un entero positivo, se recibió %q", raw)
		}
		cfg.MaxMatrixDim = value
	}

	if raw := os.Getenv("REQUEST_TIMEOUT"); raw != "" {
		value, err := time.ParseDuration(raw)
		if err != nil || value <= 0 {
			return Config{}, fmt.Errorf("REQUEST_TIMEOUT debe ser una duración positiva, se recibió %q", raw)
		}
		cfg.RequestTimeout = value
	}

	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
