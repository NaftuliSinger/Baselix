package main

import (
	"context"
	"log"

	"baselix/internal/config"
	"baselix/internal/db"
	"baselix/internal/models"
	"baselix/internal/router"
	"baselix/internal/utils"
)

func main() {
	config.Init()

	db.Init(config.Cfg)

	// insert sample users

	sampleUsers := []models.User{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "Bob", Email: "bob@example.com"},
	}

	_, err := db.DB.NewInsert().Model(&sampleUsers).Exec(context.Background())

	if err != nil {
		log.Fatal("Failed to insert sample users:", err)
	}

	utils.Debug("Sample users inserted", true)

	r := router.New()

	r.Run(":" + config.Cfg.AppPort)
	utils.Debug("Server started on port "+config.Cfg.AppPort, true)
}
