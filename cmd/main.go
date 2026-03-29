package main

import (
	"baselix/internal/config"
	"baselix/internal/db"
	"baselix/internal/router"
	"baselix/internal/utils"

	clerk "github.com/clerk/clerk-sdk-go/v2"
)

func main() {
	// Load ENV config first
	config.Init()

	// Initialize plans config once at app startup
	if err := config.InitPlans("plans.json"); err != nil {
		panic(err)
	}

	// Set Clerk secret key for authentication
	clerk.SetKey(config.Cfg.ClerkSecretKey)

	// Initialize database connection
	db.Init(config.Cfg)

	// Create and start the router
	r := router.New()
	r.Run("localhost:" + config.Cfg.AppPort)
	// Log server start
	utils.Debug("Server started on port "+config.Cfg.AppPort, true)
}
