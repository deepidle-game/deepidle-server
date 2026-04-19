package handlers

import (
	"context"
	"time"

	"deepidle-server/config"
	"deepidle-server/database"
	"deepidle-server/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Signup(c *fiber.Ctx) error {
	var req AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	collUsers := database.DB.Collection("users")
	
	// Check if user exists
	var existingUser models.User
	err := collUsers.FindOne(context.TODO(), bson.M{"username": req.Username}).Decode(&existingUser)
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Username already exists"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not hash password"})
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
	}

	result, err := collUsers.InsertOne(context.TODO(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create user"})
	}

	userID := result.InsertedID.(primitive.ObjectID)

	// Create character
	collChars := database.DB.Collection("characters")
	char := models.Character{
		UserID:          userID,
		Name:            req.Username, 
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
	_, _ = collChars.InsertOne(context.TODO(), char)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User registered successfully"})
}

func Signin(c *fiber.Ctx) error {
	var req AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	collUsers := database.DB.Collection("users")
	var user models.User
	err := collUsers.FindOne(context.TODO(), bson.M{"username": req.Username}).Decode(&user)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Generate JWT
	secret := config.GetJWTSecret()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID.Hex(),
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error creating token"})
	}

	return c.JSON(fiber.Map{"token": t})
}
