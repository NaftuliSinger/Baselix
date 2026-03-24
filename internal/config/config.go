package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv       string
	AppPort      string
	Debug        bool
	DB_DSN       string
	CreateTables bool
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
		AppEnv:       getEnv("APP_ENV", "development"),
		AppPort:      getEnv("APP_PORT", "8080"),
		Debug:        debug,
		DB_DSN:       getEnv("DB_DSN", ""),
		CreateTables: getEnv("CREATE_TABLES", "false") == "true",
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
