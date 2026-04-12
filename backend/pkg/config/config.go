package config

import (
	"context"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

type Config struct {
	Port                    string
	Environment             string
	FirebaseProjectID       string
	FirebaseCredentialsPath string
	LogLevel                string
	CORSAllowedOrigins      []string
	AlphaVantageAPIKey      string
}

func Load() *Config {
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:4200,http://localhost:3000")
	origins := strings.Split(corsOrigins, ",")

	return &Config{
		Port:                    getEnv("PORT", "8080"),
		Environment:             getEnv("ENVIRONMENT", "development"),
		FirebaseProjectID:       getEnv("FIREBASE_PROJECT_ID", ""),
		FirebaseCredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", ""),
		LogLevel:                getEnv("LOG_LEVEL", "info"),
		CORSAllowedOrigins:      origins,
		AlphaVantageAPIKey:      getEnv("ALPHA_VANTAGE_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func InitializeFirebase(ctx context.Context) (*firebase.App, error) {
	var app *firebase.App
	var err error

	cfg := Load()

	// Configure Firebase options
	var opts []option.ClientOption

	if cfg.FirebaseCredentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.FirebaseCredentialsPath))
	}

	// Configure Firebase app config
	firebaseConfig := &firebase.Config{
		ProjectID: cfg.FirebaseProjectID,
	}

	if cfg.FirebaseCredentialsPath != "" {
		app, err = firebase.NewApp(ctx, firebaseConfig, opts...)
	} else {
		// Use default credentials for local development with emulator or Google Cloud environment
		app, err = firebase.NewApp(ctx, firebaseConfig)
	}

	if err != nil {
		return nil, err
	}

	return app, nil
}
