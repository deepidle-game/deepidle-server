package handlers

import (
	"context"

	"deepidle-server/database"
	"deepidle-server/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetInventory(c *fiber.Ctx) error {
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

	return c.JSON(fiber.Map{"inventory": char.Inventory})
}

type Requirement struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

type UpgradeOption struct {
	TargetItem string        `json:"target_item"`
	BaseLevel  int           `json:"base_level"`
	Materials  []Requirement `json:"materials"`
}

func GetUpgradeOptions(c *fiber.Ctx) error {
	// Load upgrade config from database
	collConfigs := database.DB.Collection("configs")
	var gameConfig models.GameConfig
	err := collConfigs.FindOne(context.TODO(), bson.M{"config_id": "main_config"}).Decode(&gameConfig)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load game config"})
	}

	return c.JSON(fiber.Map{
		"upgrades": gameConfig.Upgrades,
	})
}

type UpgradeRequest struct {
	ItemID string `json:"item_id"`
}

func UpgradeItem(c *fiber.Ctx) error {
	userIDStr := c.Locals("userID").(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var req UpgradeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	collChars := database.DB.Collection("characters")
	var char models.Character
	err = collChars.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&char)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Character not found"})
	}

	// Find the item
	itemIndex := -1
	for i, item := range char.Inventory {
		if item.ItemID == req.ItemID {
			itemIndex = i
			break
		}
	}

	if itemIndex == -1 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Item not found in inventory"})
	}

	targetLevel := char.Inventory[itemIndex].Level

	// Load upgrade config from database
	collConfigs := database.DB.Collection("configs")
	var gameConfig models.GameConfig
	err = collConfigs.FindOne(context.TODO(), bson.M{"config_id": "main_config"}).Decode(&gameConfig)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load game config"})
	}

	upgradeConfig, exists := gameConfig.Upgrades[req.ItemID]
	if !exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "This item cannot be upgraded"})
	}

	// Calculate required materials based on target level
	// Formula: material.quantity * targetLevel
	required := []struct {
		ItemID   string
		Quantity int
	}{}
	for _, mat := range upgradeConfig.Materials {
		required = append(required, struct {
			ItemID   string
			Quantity int
		}{
			ItemID:   mat.ItemID,
			Quantity: mat.Quantity * targetLevel,
		})
	}

	// Check if user has all materials
	for _, reqMat := range required {
		found := false
		for _, invItem := range char.Inventory {
			if invItem.ItemID == reqMat.ItemID && invItem.Quantity >= reqMat.Quantity {
				found = true
				break
			}
		}
		if !found {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":    "Not enough materials",
				"required": required,
			})
		}
	}

	// Deduct materials
	for _, reqMat := range required {
		for i, invItem := range char.Inventory {
			if invItem.ItemID == reqMat.ItemID {
				char.Inventory[i].Quantity -= reqMat.Quantity
				break
			}
		}
	}

	// Remove items with 0 quantity to free inventory slots
	newInventory := []models.Item{}
	for _, item := range char.Inventory {
		if item.Quantity > 0 {
			newInventory = append(newInventory, item)
		}
	}
	char.Inventory = newInventory

	char.Inventory[itemIndex].Level += 1

	_, err = collChars.UpdateOne(context.TODO(), bson.M{"user_id": userID}, bson.M{"$set": bson.M{"inventory": char.Inventory}})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to upgrade item"})
	}

	return c.JSON(fiber.Map{"message": "Item upgraded", "item": char.Inventory[itemIndex]})
}
