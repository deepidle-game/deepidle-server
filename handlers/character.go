package handlers

import (
	"context"
	"time"

	"deepidle-server/claims"
	"deepidle-server/config"
	"deepidle-server/database"
	"deepidle-server/models"
	"deepidle-server/state"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetCharacter(c *fiber.Ctx) error {
	userIDStr := c.Locals("userID").(string)
	activeCharID := c.Locals("characterID").(string)
	
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	charID, err := primitive.ObjectIDFromHex(activeCharID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid character ID"})
	}

	collChars := database.DB.Collection("characters")
	var char models.Character
	err = collChars.FindOne(context.TODO(), bson.M{"_id": charID, "user_id": userID}).Decode(&char)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Character not found"})
	}

	return c.JSON(fiber.Map{
		"id":             char.ID.Hex(),
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
	activeCharID := c.Locals("characterID").(string)

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	charID, err := primitive.ObjectIDFromHex(activeCharID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid character ID"})
	}

	var req ActionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	collChars := database.DB.Collection("characters")
	_, err = collChars.UpdateOne(
		context.TODO(),
		bson.M{"_id": charID, "user_id": userID},
		bson.M{"$set": bson.M{
			"current_action":   req.Action,
			"action_started_at": time.Now().Unix(),
		}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update action"})
	}

	state.UpdatePlayerState(userIDStr, username, req.Action)

	return c.JSON(fiber.Map{"message": "Action updated", "current_action": req.Action})
}

func GetOnlinePlayers(c *fiber.Ctx) error {
	players := state.GetOnlinePlayers()
	return c.JSON(fiber.Map{"online_players": players})
}

func ClaimResources(c *fiber.Ctx) error {
	userIDStr := c.Locals("userID").(string)
	activeCharID := c.Locals("characterID").(string)

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	charID, err := primitive.ObjectIDFromHex(activeCharID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid character ID"})
	}

	collChars := database.DB.Collection("characters")
	var char models.Character
	err = collChars.FindOne(context.TODO(), bson.M{"_id": charID, "user_id": userID}).Decode(&char)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Character not found"})
	}

	collConfigs := database.DB.Collection("configs")
	var gameConfig models.GameConfig
	err = collConfigs.FindOne(context.TODO(), bson.M{"config_id": "main_config"}).Decode(&gameConfig)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load game config"})
	}

	actionCfg, ok := gameConfig.Actions[char.CurrentAction]
	if !ok {
		return c.JSON(fiber.Map{"message": "Current action does not yield resources", "gained": nil})
	}

	timePassed := time.Now().Unix() - char.ActionStartedAt

	valid, msg := claims.ValidateClaim(&char, actionCfg, timePassed)
	if !valid {
		return c.JSON(fiber.Map{"message": msg, "gained": nil})
	}

	resourceCount, resourceType := claims.CalculateResources(&char, actionCfg, timePassed)
	if resourceCount <= 0 {
		return c.JSON(fiber.Map{"message": "Not enough resources gathered", "gained": nil})
	}

	stacked, newSlot, newInventory := claims.AddResourceToInventory(
		char.Inventory, resourceType, resourceCount, char.MaxInventorySlots,
	)

	if !stacked && !newSlot {
		currentItems := make([]string, len(char.Inventory))
		for i, item := range char.Inventory {
			currentItems[i] = item.ItemID
		}
		return c.JSON(fiber.Map{
			"message":           "Inventory full. No empty slots for new item type.",
			"gained":           nil,
			"inventory_items":   currentItems,
			"new_item_type":    resourceType,
			"inventory_slots":   len(char.Inventory),
			"max_slots":        char.MaxInventorySlots,
			"suggestion":       "Deposit an item to free a slot, or stack with existing resources.",
		})
	}

	char.Inventory = newInventory
	_, err = collChars.UpdateOne(
		context.TODO(),
		bson.M{"_id": charID, "user_id": userID},
		bson.M{"$set": bson.M{
			"inventory":         char.Inventory,
			"action_started_at": time.Now().Unix(),
		}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to claim resources"})
	}

	msg = "Resources claimed and stacked"
	if newSlot {
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
	activeCharID := c.Locals("characterID").(string)

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	charID, err := primitive.ObjectIDFromHex(activeCharID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid character ID"})
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
		bson.M{"_id": charID, "user_id": userID},
		bson.M{"$set": bson.M{"name": req.Name}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update name"})
	}

	return c.JSON(fiber.Map{"message": "Character name updated", "name": req.Name})
}

func ListCharacters(c *fiber.Ctx) error {
	userIDStr := c.Locals("userID").(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	collChars := database.DB.Collection("characters")
	cursor, err := collChars.Find(context.TODO(), bson.M{"user_id": userID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch characters"})
	}

	var characters []models.Character
	if err = cursor.All(context.TODO(), &characters); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode characters"})
	}

	charsJSON := make([]fiber.Map, len(characters))
	for i, char := range characters {
		charsJSON[i] = fiber.Map{
			"id":             char.ID.Hex(),
			"name":           char.Name,
			"level":          char.Level,
			"current_action": char.CurrentAction,
		}
	}

	return c.JSON(fiber.Map{"characters": charsJSON})
}

type CreateCharacterRequest struct {
	Name string `json:"name"`
}

func CreateCharacter(c *fiber.Ctx) error {
	userIDStr := c.Locals("userID").(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var req CreateCharacterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(req.Name) < 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name too short"})
	}

	collChars := database.DB.Collection("characters")
	char := models.Character{
		UserID:          userID,
		Name:            req.Name,
		Level:           1,
		CurrentAction:   "Idle",
		ActionStartedAt: time.Now().Unix(),
		MaxInventorySlots: 5,
		Inventory: []models.Item{
			{ItemID: "wooden_axe", Level: 1, Quantity: 1},
			{ItemID: "wooden_pickaxe", Level: 1, Quantity: 1},
			{ItemID: "wooden_sword", Level: 1, Quantity: 1},
		},
	}

	result, err := collChars.InsertOne(context.TODO(), char)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create character"})
	}

	characterID := result.InsertedID.(primitive.ObjectID)

	return c.JSON(fiber.Map{
		"message":    "Character created",
		"character": fiber.Map{
			"id":    characterID.Hex(),
			"name":  char.Name,
			"level": char.Level,
		},
	})
}

type SelectCharacterRequest struct {
	CharacterID string `json:"character_id"`
}

func SelectCharacter(c *fiber.Ctx) error {
	userIDStr := c.Locals("userID").(string)
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var req SelectCharacterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	charID, err := primitive.ObjectIDFromHex(req.CharacterID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid character ID"})
	}

	collChars := database.DB.Collection("characters")
	var char models.Character
	err = collChars.FindOne(context.TODO(), bson.M{"_id": charID, "user_id": userID}).Decode(&char)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Character not found"})
	}

	collUsers := database.DB.Collection("users")
	_, err = collUsers.UpdateOne(
		context.TODO(),
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"active_character_id": charID}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update active character"})
	}

	username := c.Locals("username").(string)
	secret := config.GetJWTSecret()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userIDStr,
		"username": username,
		"character_id": char.ID.Hex(),
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	newToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create token"})
	}

	return c.JSON(fiber.Map{
		"message":     "Character selected",
		"character_id": char.ID.Hex(),
		"name":        char.Name,
		"token":       newToken,
	})
}

