package database

import (
	"context"
	"log"

	"deepidle-server/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SeedDatabase() {
	coll := DB.Collection("configs")

	configID := "main_config"
	filter := bson.M{"config_id": configID}

	defaults := models.GameConfig{
		ConfigID: configID,
		BaseSlots: 5,
		Actions: map[string]models.ActionConfig{
			"cutting_wood": {
				RequiredTool:  "wooden_axe",
				DropItem:      "wood",
				BaseTimeSec:   10,
				BaseAmount:    1,
				BonusPerLevel: 0.20, // 20% bonus drops per level
			},
			"mining_rocks": {
				RequiredTool:  "wooden_pickaxe",
				DropItem:      "stone",
				BaseTimeSec:   15,
				BaseAmount:    1,
				BonusPerLevel: 0.20,
			},
			"fighting_monsters": {
				RequiredTool:  "wooden_sword",
				DropItem:      "gold",
				BaseTimeSec:   20,
				BaseAmount:    5,
				BonusPerLevel: 0.20,
			},
		},
		Upgrades: map[string]models.UpgradeConfig{
			"wooden_axe": {
				BaseLevel: 1,
				Materials: []models.UpgradeMaterial{
					{ItemID: "wood", Quantity: 5},
					{ItemID: "gold", Quantity: 10},
				},
			},
			"wooden_pickaxe": {
				BaseLevel: 1,
				Materials: []models.UpgradeMaterial{
					{ItemID: "stone", Quantity: 5},
					{ItemID: "wood", Quantity: 5},
					{ItemID: "gold", Quantity: 10},
				},
			},
			"wooden_sword": {
				BaseLevel: 1,
				Materials: []models.UpgradeMaterial{
					{ItemID: "wood", Quantity: 5},
					{ItemID: "gold", Quantity: 10},
				},
			},
		},
	}

	update := bson.M{"$setOnInsert": defaults}
	opts := options.Update().SetUpsert(true)

	_, err := coll.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		log.Printf("Failed to seed database: %v", err)
	} else {
		log.Println("Database seeded with default configurations")
	}
}
