package handlers

import (
	"context"

	"deepidle-server/database"
	"deepidle-server/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetGlobalStorage(c *fiber.Ctx) error {
	collStorage := database.DB.Collection("global_storage")
	cursor, err := collStorage.Find(context.TODO(), bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch storage"})
	}
	defer cursor.Close(context.TODO())

	var items []models.GlobalStorageItem
	if err := cursor.All(context.TODO(), &items); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not parse storage"})
	}

	return c.JSON(fiber.Map{"storage": items})
}

type StorageActionRequest struct {
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

func DepositStorage(c *fiber.Ctx) error {
	userIDStr := c.Locals("userID").(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var req StorageActionRequest
	if err := c.BodyParser(&req); err != nil || req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body or quantity"})
	}

	// Fetch character
	collChars := database.DB.Collection("characters")
	var char models.Character
	if err := collChars.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&char); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Character not found"})
	}

	// Find item in inventory
	foundIdx := -1
	itemLevel := 1
	for i, item := range char.Inventory {
		if item.ItemID == req.ItemID {
			foundIdx = i
			itemLevel = item.Level
			break
		}
	}

	if foundIdx == -1 || char.Inventory[foundIdx].Quantity < req.Quantity {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Not enough items in inventory"})
	}

	// Deduct from inventory
	char.Inventory[foundIdx].Quantity -= req.Quantity
	if char.Inventory[foundIdx].Quantity == 0 {
		// Remove item entirely if quantity reaches 0
		char.Inventory = append(char.Inventory[:foundIdx], char.Inventory[foundIdx+1:]...)
	}

	// Update Character DB
	_, err = collChars.UpdateOne(context.TODO(), bson.M{"user_id": userID}, bson.M{"$set": bson.M{"inventory": char.Inventory}})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update inventory"})
	}

	// Add to global storage (with level for tools)
	collStorage := database.DB.Collection("global_storage")
	opts := options.Update().SetUpsert(true)
	_, err = collStorage.UpdateOne(
		context.TODO(),
		bson.M{"item_id": req.ItemID},
		bson.M{
			"$inc":  bson.M{"quantity": req.Quantity},
			"$set":  bson.M{"level": itemLevel},
		},
		opts,
	)

	return c.JSON(fiber.Map{"message": "Deposited successfully", "inventory": char.Inventory})
}

func WithdrawStorage(c *fiber.Ctx) error {
	userIDStr := c.Locals("userID").(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var req StorageActionRequest
	if err := c.BodyParser(&req); err != nil || req.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body or quantity"})
	}

	collStorage := database.DB.Collection("global_storage")

	// Check if in storage and has enough quantity
	var storageItem models.GlobalStorageItem
	if err := collStorage.FindOne(context.TODO(), bson.M{"item_id": req.ItemID}).Decode(&storageItem); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Item not found in global storage"})
	}

	if storageItem.Quantity < req.Quantity {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Not enough items in global storage"})
	}

	// Fetch character
	collChars := database.DB.Collection("characters")
	var char models.Character
	if err := collChars.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&char); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Character not found"})
	}

	// Check if item already exists in inventory (for tools, there's only 1)
	foundIdx := -1
	for i, item := range char.Inventory {
		if item.ItemID == req.ItemID {
			foundIdx = i
			break
		}
	}

	if foundIdx != -1 {
		// Item exists - this shouldn't happen for tools (you can't have 2 axes)
		// But for stackable resources, increase quantity
		char.Inventory[foundIdx].Quantity += req.Quantity
	} else {
		// New item
		if len(char.Inventory) >= char.MaxInventorySlots {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Inventory full. Cannot withdraw new item."})
		}
		// Use stored level if available, otherwise default to 1
		level := storageItem.Level
		if level == 0 {
			level = 1
		}
		char.Inventory = append(char.Inventory, models.Item{
			ItemID:   req.ItemID,
			Level:    level,
			Quantity: req.Quantity,
		})
	}

	// Deduct from global storage
	if storageItem.Quantity == req.Quantity {
		// Delete record if empty
		collStorage.DeleteOne(context.TODO(), bson.M{"item_id": req.ItemID})
	} else {
		collStorage.UpdateOne(context.TODO(), bson.M{"item_id": req.ItemID}, bson.M{"$inc": bson.M{"quantity": -req.Quantity}})
	}

	// Update Character DB
	_, err = collChars.UpdateOne(context.TODO(), bson.M{"user_id": userID}, bson.M{"$set": bson.M{"inventory": char.Inventory}})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update inventory"})
	}

	return c.JSON(fiber.Map{"message": "Withdrawn successfully", "inventory": char.Inventory})
}
