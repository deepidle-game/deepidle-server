package handlers

import (
	"context"

	"deepidle-server/database"
	"deepidle-server/inventory"
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

	collChars := database.DB.Collection("characters")
	var char models.Character
	if err := collChars.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&char); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Character not found"})
	}

	foundIdx := inventory.FindItemIndex(char.Inventory, req.ItemID)
	if foundIdx == -1 || char.Inventory[foundIdx].Quantity < req.Quantity {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Not enough items in inventory"})
	}

	itemLevel := char.Inventory[foundIdx].Level
	char.Inventory = inventory.DeductMaterials(char.Inventory, []struct {
		ItemID   string
		Quantity int
	}{{ItemID: req.ItemID, Quantity: req.Quantity}})

	_, err = collChars.UpdateOne(context.TODO(), bson.M{"user_id": userID}, bson.M{"$set": bson.M{"inventory": char.Inventory}})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update inventory"})
	}

	collStorage := database.DB.Collection("global_storage")
	opts := options.Update().SetUpsert(true)
	_, err = collStorage.UpdateOne(
		context.TODO(),
		bson.M{"item_id": req.ItemID},
		bson.M{
			"$inc": bson.M{"quantity": req.Quantity},
			"$set": bson.M{"level": itemLevel},
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

	var storageItem models.GlobalStorageItem
	if err := collStorage.FindOne(context.TODO(), bson.M{"item_id": req.ItemID}).Decode(&storageItem); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Item not found in global storage"})
	}

	if storageItem.Quantity < req.Quantity {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Not enough items in global storage"})
	}

	collChars := database.DB.Collection("characters")
	var char models.Character
	if err := collChars.FindOne(context.TODO(), bson.M{"user_id": userID}).Decode(&char); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Character not found"})
	}

	foundIdx := inventory.FindItemIndex(char.Inventory, req.ItemID)
	if foundIdx != -1 {
		char.Inventory[foundIdx].Quantity += req.Quantity
	} else {
		if len(char.Inventory) >= char.MaxInventorySlots {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Inventory full. Cannot withdraw new item."})
		}
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

	if storageItem.Quantity == req.Quantity {
		collStorage.DeleteOne(context.TODO(), bson.M{"item_id": req.ItemID})
	} else {
		collStorage.UpdateOne(context.TODO(), bson.M{"item_id": req.ItemID}, bson.M{"$inc": bson.M{"quantity": -req.Quantity}})
	}

	_, err = collChars.UpdateOne(context.TODO(), bson.M{"user_id": userID}, bson.M{"$set": bson.M{"inventory": char.Inventory}})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update inventory"})
	}

	return c.JSON(fiber.Map{"message": "Withdrawn successfully", "inventory": char.Inventory})
}
