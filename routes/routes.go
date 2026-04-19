package routes

import (
	"deepidle-server/handlers"
	"deepidle-server/middleware"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	api := app.Group("/api")

	// Auth
	auth := api.Group("/auth")
	auth.Post("/signup", handlers.Signup)
	auth.Post("/signin", handlers.Signin)

	// Protected Routes
	protected := api.Group("/", middleware.Protected())

	// Character
	character := protected.Group("/character")
	character.Get("/detail", handlers.GetCharacter)
	character.Post("/action", handlers.UpdateAction)
	character.Post("/claim", handlers.ClaimResources)
	character.Patch("/name", handlers.UpdateCharacterName)

	// Character Management (multi-character)
	characters := protected.Group("/characters")
	characters.Get("/list", handlers.ListCharacters)
	characters.Post("/create", handlers.CreateCharacter)
	characters.Post("/select", handlers.SelectCharacter)

	// Inventory
	inventory := protected.Group("/inventory")
	inventory.Get("/", handlers.GetInventory)
	inventory.Get("/upgrade-options", handlers.GetUpgradeOptions)
	inventory.Post("/upgrade", handlers.UpgradeItem)

	// Players
	players := protected.Group("/players")
	players.Get("/online", handlers.GetOnlinePlayers)

	// Global Storage
	storage := protected.Group("/storage")
	storage.Get("/", handlers.GetGlobalStorage)
	storage.Post("/deposit", handlers.DepositStorage)
	storage.Post("/withdraw", handlers.WithdrawStorage)
}
