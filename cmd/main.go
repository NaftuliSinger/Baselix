package main

import (
	"baselix/internal/config"
	"baselix/internal/db"
	"baselix/internal/router"
	"baselix/internal/utils"

	clerk "github.com/clerk/clerk-sdk-go/v2"
)

func main() {
	config.Init()
	clerk.SetKey(config.Cfg.ClerkSecretKey)

	db.Init(config.Cfg)

	utils.Debug("Sample users inserted", true)

	r := router.New()

	r.Run(":" + config.Cfg.AppPort)
	utils.Debug("Server started on port "+config.Cfg.AppPort, true)
}
