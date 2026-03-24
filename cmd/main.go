package main

import (
	"baselix/internal/config"
	"baselix/internal/db"
	"baselix/internal/router"
	"baselix/internal/utils"
	"fmt"

	clerk "github.com/clerk/clerk-sdk-go/v2"
)

func main() {
	// Load config first
	config.Init()

	// Initialize plans config once at app startup
	if err := config.InitPlans("plans.json"); err != nil {
		panic(err)
	}

	// Access any plan anywhere
	freePlan, err := config.GetPlan("free")
	if err != nil {
		panic(err)
	}

	freeRecordsLimit := config.GetPlanLimit("free", "records", 100)
	fmt.Println("Free Plan Records Limit:", freeRecordsLimit)

	fmt.Println("Free Plan Limits:", freePlan.Limits)
	fmt.Println("Free Plan Features:", freePlan.Features)

	// Another example
	proPlan, _ := config.GetPlan("pro")
	fmt.Println("Pro Plan Projects Limit:", proPlan.Limits["projects"])

	clerk.SetKey(config.Cfg.ClerkSecretKey)

	db.Init(config.Cfg)

	r := router.New()

	r.Run(":" + config.Cfg.AppPort)
	utils.Debug("Server started on port "+config.Cfg.AppPort, true)
}
