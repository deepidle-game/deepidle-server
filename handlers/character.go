package handlers

import (
	"context"
	"time"

	"deepidle-server/database"
	"deepidle-server/models"
	"deepidle-server/state"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetCharacter(c *fiber.Ctx) error {
	userIDStr := c.Locals("userID").(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	collChars := database.DB.Collection("characters")
	var char models.Character
	err = collChars.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&char)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Character not found"})
	}

	return c.JSON(fiber.Map{
		"name":           char.Name,
		"level":          char.Level,
		"current_action": char.CurrentAction,
	})
}

type ActionRequest struct {
	Action string `json:"action"`
}

func UpdateAction(c *fiber.Ctx) error {
	userIDStr := c.Locals("userID").(string)
	username := c.Locals("username").(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var req ActionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	collChars := database.DB.Collection("characters")
	_, err = collChars.UpdateOne(
		context.TODO(), 
		bson.M{"user_id": userID}, 
		bson.M{"$set": bson.M{
			"current_action": req.Action, 
			"action_started_at": time.Now().Unix(),
		}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update action"})
	}

	// Update in-memory state
	state.UpdatePlayerState(userIDStr, username, req.Action)

	return c.JSON(fiber.Map{"message": "Action updated", "current_action": req.Action})
}

func GetOnlinePlayers(c *fiber.Ctx) error {
	// Directly from in-memory state
	players := state.GetOnlinePlayers()
	return c.JSON(fiber.Map{"online_players": players})
}

func ClaimResources(c *fiber.Ctx) error {
	userIDStr := c.Locals("userID").(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	collChars := database.DB.Collection("characters")
	var char models.Character
	err = collChars.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&char)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Character not found"})
	}

	if char.CurrentAction == "Idle" || char.ActionStartedAt == 0 {
		return c.JSON(fiber.Map{"message": "No active action to claim resources from", "gained": nil})
	}

	// Fetch Config
	collConfigs := database.DB.Collection("configs")
	var gameConfig models.GameConfig
	err = collConfigs.FindOne(context.TODO(), bson.M{"config_id": "main_config"}).Decode(&gameConfig)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load game config"})
	}

	actionCfg, ok := gameConfig.Actions[char.CurrentAction]
	if !ok {
		// Default to not yielding resources
		return c.JSON(fiber.Map{"message": "Current action does not yield resources", "gained": nil})
	}

	timePassed := time.Now().Unix() - char.ActionStartedAt

	// 1. Determine tool level
	toolLevel := 0
	for _, item := range char.Inventory {
		if item.ItemID == actionCfg.RequiredTool {
			toolLevel = item.Level
			break
		}
	}

	if toolLevel == 0 {
		return c.JSON(fiber.Map{
			"message":       "You don't have the required tool for this action",
			"required_tool":  actionCfg.RequiredTool,
			"gained":        nil,
		})
	}

	// 2. Calculate resource using bonus percentage
	cycles := int(timePassed / actionCfg.BaseTimeSec)
	if cycles <= 0 {
		return c.JSON(fiber.Map{"message": "Not enough time passed to claim resources", "gained": nil})
	}

	totalBaseYield := float64(cycles * actionCfg.BaseAmount)
	multiplier := 1.0 + (float64(toolLevel-1) * actionCfg.BonusPerLevel)
	resourceCount := int(totalBaseYield * multiplier)

	if resourceCount <= 0 {
		return c.JSON(fiber.Map{"message": "Not enough resources gathered", "gained": nil})
	}

	resourceType := actionCfg.DropItem

	// Add resource to inventory
	found := false
	for i, item := range char.Inventory {
		if item.ItemID == resourceType {
			char.Inventory[i].Quantity += resourceCount
			found = true
			break
		}
	}

	if !found {
		// Enforce inventory limit for new items
		if len(char.Inventory) >= char.MaxInventorySlots {
			// List current items to help user understand
			currentItems := []string{}
			for _, item := range char.Inventory {
				currentItems = append(currentItems, item.ItemID)
			}
			return c.JSON(fiber.Map{
				"message":          "Inventory full. No empty slots for new item type.",
				"gained":          nil,
				"inventory_items":  currentItems,
				"new_item_type":   resourceType,
				"inventory_slots": len(char.Inventory),
				"max_slots":       char.MaxInventorySlots,
				"suggestion":      "Deposit an item to free a slot, or stack with existing resources.",
			})
		}
		char.Inventory = append(char.Inventory, models.Item{
			ItemID:   resourceType,
			Level:    1,
			Quantity: resourceCount,
		})
	}

	// Update DB (both inventory and reset action timer)
	_, err = collChars.UpdateOne(
		context.TODO(),
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{
			"inventory": char.Inventory,
			"action_started_at": time.Now().Unix(),
		}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to claim resources"})
	}

	msg := "Resources claimed and stacked"
	if !found {
		msg = "Resources claimed (new slot used)"
	}
	return c.JSON(fiber.Map{
		"message": msg,
		"gained": map[string]int{
			resourceType: resourceCount,
		},
	})
}

type NameRequest struct {
	Name string `json:"name"`
}

func UpdateCharacterName(c *fiber.Ctx) error {
	userIDStr := c.Locals("userID").(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var req NameRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(req.Name) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name too short"})
	}

	collChars := database.DB.Collection("characters")
	_, err = collChars.UpdateOne(
		context.TODO(), 
		bson.M{"user_id": userID}, 
		bson.M{"$set": bson.M{"name": req.Name}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update name"})
	}

	return c.JSON(fiber.Map{"message": "Character name updated", "name": req.Name})
}

