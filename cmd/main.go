package main

import (
	"baselix/internal/config"
	"baselix/internal/db"
	"baselix/internal/router"
	"baselix/internal/utils"
)

func main() {
	config.Init()

	db.Init(config.Cfg)

	utils.Debug("Sample users inserted", true)

	r := router.New()

	r.Run(":" + config.Cfg.AppPort)
	utils.Debug("Server started on port "+config.Cfg.AppPort, true)
}
