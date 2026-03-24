package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// PlanConfig represents limits and features of a plan
type PlanConfig struct {
	Limits   map[string]int `json:"limits"`
	Features []string       `json:"features"`
}

// Plans represents all plans
type Plans struct {
	Plans map[string]PlanConfig `json:"plans"`
}

// AllPlans is the global variable accessible everywhere
var AllPlans *Plans

// InitPlans loads plan config from JSON and sets the global AllPlans
func InitPlans(filePath string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var plans Plans
	if err := json.Unmarshal(file, &plans); err != nil {
		return err
	}

	AllPlans = &plans
	return nil
}

// GetPlan returns a specific plan config by name from the global AllPlans
func GetPlan(name string) (*PlanConfig, error) {
	if AllPlans == nil {
		return nil, fmt.Errorf("plans not initialized")
	}

	plan, ok := AllPlans.Plans[name]
	if !ok {
		return nil, fmt.Errorf("plan %s not found", name)
	}
	return &plan, nil
}

// GetPlanLimit returns a specific limit value from a plan.
// If the plan or limit does not exist, returns the provided default value.
func GetPlanLimit(planName string, limitKey string, defaultValue int) int {
	plan, err := GetPlan(planName)
	if err != nil {
		return defaultValue
	}

	value, ok := plan.Limits[limitKey]
	if !ok {
		return defaultValue
	}
	return value
}
