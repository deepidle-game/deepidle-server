package handlers

import (
	"context"

	"deepidle-server/database"
	"deepidle-server/inventory"
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

	itemIndex := inventory.FindItemIndex(char.Inventory, req.ItemID)
	if itemIndex == -1 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Item not found in inventory"})
	}

	targetLevel := char.Inventory[itemIndex].Level

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

	if !inventory.HasMaterials(char.Inventory, required) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Not enough materials",
			"required": required,
		})
	}

	char.Inventory = inventory.DeductMaterials(char.Inventory, required)
	char.Inventory[itemIndex].Level += 1

	_, err = collChars.UpdateOne(context.TODO(), bson.M{"user_id": userID}, bson.M{"$set": bson.M{"inventory": char.Inventory}})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to upgrade item"})
	}

	return c.JSON(fiber.Map{"message": "Item upgraded", "item": char.Inventory[itemIndex]})
}

func GetUpgradeOptions(c *fiber.Ctx) error {
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
