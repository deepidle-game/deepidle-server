package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Item struct {
	ItemID   string `bson:"item_id" json:"item_id"`
	Level    int    `bson:"level" json:"level"`
	Quantity int    `bson:"quantity" json:"quantity"`
}

type Character struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	Name            string             `bson:"name" json:"name"`
	Level           int                `bson:"level" json:"level"`
	CurrentAction   string             `bson:"current_action" json:"current_action"`
	ActionStartedAt   int64              `bson:"action_started_at" json:"action_started_at"`
	Inventory         []Item             `bson:"inventory" json:"inventory"`
	MaxInventorySlots int                `bson:"max_inventory_slots" json:"max_inventory_slots"`
}
