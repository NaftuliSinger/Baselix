package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv              string
	Localhost           bool
	AppPort             string
	Debug               bool
	DB_DSN              string
	CreateTables        bool
	ClerkSecretKey      string
	ClerkPublishableKey string
	ClerkPublicKey      string
	ROrigin             string
	ClerkWebhookSecret  string
}

var Cfg *Config

func Init() {
	// load .env (ignore error in production)
	_ = godotenv.Load()

	debug, err := strconv.ParseBool(getEnv("DEBUG", "false"))
	if err != nil {
		debug = false
	}

	Cfg = &Config{
		AppEnv:              getEnv("APP_ENV", "development"),
		Localhost:           getEnv("LOCALHOST", "false") == "true",
		AppPort:             getEnv("APP_PORT", "8080"),
		Debug:               debug,
		DB_DSN:              getEnv("DB_DSN", ""),
		CreateTables:        getEnv("CREATE_TABLES", "false") == "true",
		ClerkSecretKey:      getEnv("CLERK_SECRET_KEY", ""),
		ClerkPublishableKey: getEnv("CLERK_PUBLISHABLE_KEY", ""),
		ClerkPublicKey:      getEnv("CLERK_PUBLIC_KEY", ""),
		ROrigin:             getEnv("R_ORIGIN", "http://localhost:8080"),
		ClerkWebhookSecret:  getEnv("CLERK_WEBHOOK_SECRET", ""),
	}

	log.Println("Config loaded")
}

func getEnv(key, fallback string) string {
	val, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return val
}
