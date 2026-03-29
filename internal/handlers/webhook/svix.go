package webhookHandlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"baselix/internal/config"
	"baselix/internal/db"
	"baselix/internal/models"
	"baselix/internal/utils"

	"github.com/gin-gonic/gin"
	svix "github.com/svix/svix-webhooks/go"
)

func SvixWebhook(c *gin.Context) {
	wh, err := svix.NewWebhook(config.Cfg.ClerkWebhookSecret)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	headers := c.Request.Header
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = wh.Verify(payload, headers)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	eventType := parsed["type"].(string)

	utils.Debug("Received event: "+eventType, true)

	switch eventType {
	case "user.created":
		userId := parsed["data"].(map[string]interface{})["id"].(string)
		newUser := &models.User{
			UserID: userId,
			Plan:   "free", // default plan
		}
		err := db.InsertUser(c, newUser)
		if err != nil {
			utils.Debug("Error inserting new user: "+err.Error(), true)
		}
		utils.Debug("Created new user with ID: "+userId, true)

	case "subscriptionItem.active":
		data := parsed["data"].(map[string]interface{})

		payer := data["payer"].(map[string]interface{})
		userId := payer["user_id"].(string)

		plan := data["plan"].(map[string]interface{})
		planSlug := plan["slug"].(string)

		user, err := db.SelectUserByID(c, userId)
		if err != nil {
			utils.Debug("Error fetching user for subscription update: "+err.Error(), true)
			break
		}
		user.Plan = planSlug

		err = db.UpdateUser(c, user)
		if err != nil {
			utils.Debug("Error updating user plan: "+err.Error(), true)
		}
		utils.Debug("Updated user "+userId+" to plan "+planSlug, true)

	default:
		utils.Debug("Received unhandled event type: "+eventType, true)
	}
	c.JSON(200, gin.H{})
}
