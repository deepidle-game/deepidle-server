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
		Resources: []string{"wood", "stone", "gold", "iron", "steel", "diamond"},
		Actions: map[string]models.ActionConfig{
			"cutting_wood": {
				RequiredTool:  "wooden_axe",
				DropItem:      "wood",
				BaseTimeSec:   10,
				BaseAmount:    1,
				BonusPerLevel: 0.20,
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
			"advanced_woodcutting": {
				RequiredTool:  "stone_axe",
				DropItem:      "wood",
				BaseTimeSec:   8,
				BaseAmount:    3,
				BonusPerLevel: 0.25,
			},
			"deep_mining": {
				RequiredTool:  "stone_pickaxe",
				DropItem:      "stone",
				BaseTimeSec:   12,
				BaseAmount:    3,
				BonusPerLevel: 0.25,
			},
			"hunting_goblins": {
				RequiredTool:  "stone_sword",
				DropItem:      "gold",
				BaseTimeSec:   18,
				BaseAmount:    10,
				BonusPerLevel: 0.25,
			},
			"iron_mining": {
				RequiredTool:  "iron_pickaxe",
				DropItem:      "iron",
				BaseTimeSec:   20,
				BaseAmount:    2,
				BonusPerLevel: 0.30,
			},
			"slaying_orcs": {
				RequiredTool:  "iron_sword",
				DropItem:      "gold",
				BaseTimeSec:   25,
				BaseAmount:    20,
				BonusPerLevel: 0.30,
			},
			"steel_forging": {
				RequiredTool:  "steel_pickaxe",
				DropItem:      "steel",
				BaseTimeSec:   30,
				BaseAmount:    2,
				BonusPerLevel: 0.35,
			},
			"dragon_slaying": {
				RequiredTool:  "steel_sword",
				DropItem:      "gold",
				BaseTimeSec:   40,
				BaseAmount:    50,
				BonusPerLevel: 0.35,
			},
			"gem_mining": {
				RequiredTool:  "diamond_pickaxe",
				DropItem:      "diamond",
				BaseTimeSec:   45,
				BaseAmount:    1,
				BonusPerLevel: 0.40,
			},
			"legendary_quests": {
				RequiredTool:  "diamond_sword",
				DropItem:      "gold",
				BaseTimeSec:   60,
				BaseAmount:    100,
				BonusPerLevel: 0.40,
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
			"stone_axe": {
				BaseLevel: 1,
				Materials: []models.UpgradeMaterial{
					{ItemID: "stone", Quantity: 15},
					{ItemID: "wood", Quantity: 10},
					{ItemID: "gold", Quantity: 25},
				},
			},
			"stone_pickaxe": {
				BaseLevel: 1,
				Materials: []models.UpgradeMaterial{
					{ItemID: "stone", Quantity: 20},
					{ItemID: "wood", Quantity: 15},
					{ItemID: "gold", Quantity: 30},
				},
			},
			"stone_sword": {
				BaseLevel: 1,
				Materials: []models.UpgradeMaterial{
					{ItemID: "stone", Quantity: 15},
					{ItemID: "wood", Quantity: 10},
					{ItemID: "gold", Quantity: 25},
				},
			},
			"iron_axe": {
				BaseLevel: 1,
				Materials: []models.UpgradeMaterial{
					{ItemID: "iron", Quantity: 10},
					{ItemID: "stone", Quantity: 25},
					{ItemID: "gold", Quantity: 50},
				},
			},
			"iron_pickaxe": {
				BaseLevel: 1,
				Materials: []models.UpgradeMaterial{
					{ItemID: "iron", Quantity: 15},
					{ItemID: "stone", Quantity: 30},
					{ItemID: "gold", Quantity: 60},
				},
			},
			"iron_sword": {
				BaseLevel: 1,
				Materials: []models.UpgradeMaterial{
					{ItemID: "iron", Quantity: 10},
					{ItemID: "stone", Quantity: 25},
					{ItemID: "gold", Quantity: 50},
				},
			},
			"steel_axe": {
				BaseLevel: 1,
				Materials: []models.UpgradeMaterial{
					{ItemID: "steel", Quantity: 10},
					{ItemID: "iron", Quantity: 20},
					{ItemID: "gold", Quantity: 100},
				},
			},
			"steel_pickaxe": {
				BaseLevel: 1,
				Materials: []models.UpgradeMaterial{
					{ItemID: "steel", Quantity: 15},
					{ItemID: "iron", Quantity: 25},
					{ItemID: "gold", Quantity: 120},
				},
			},
			"steel_sword": {
				BaseLevel: 1,
				Materials: []models.UpgradeMaterial{
					{ItemID: "steel", Quantity: 10},
					{ItemID: "iron", Quantity: 20},
					{ItemID: "gold", Quantity: 100},
				},
			},
			"diamond_axe": {
				BaseLevel: 1,
				Materials: []models.UpgradeMaterial{
					{ItemID: "diamond", Quantity: 5},
					{ItemID: "steel", Quantity: 15},
					{ItemID: "gold", Quantity: 200},
				},
			},
			"diamond_pickaxe": {
				BaseLevel: 1,
				Materials: []models.UpgradeMaterial{
					{ItemID: "diamond", Quantity: 8},
					{ItemID: "steel", Quantity: 20},
					{ItemID: "gold", Quantity: 250},
				},
			},
			"diamond_sword": {
				BaseLevel: 1,
				Materials: []models.UpgradeMaterial{
					{ItemID: "diamond", Quantity: 5},
					{ItemID: "steel", Quantity: 15},
					{ItemID: "gold", Quantity: 200},
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
