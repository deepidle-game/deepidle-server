package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// GlobalStorageItem represents a single item stack in the global storage pool
type GlobalStorageItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ItemID   string           `bson:"item_id" json:"item_id"`
	Quantity int              `bson:"quantity" json:"quantity"`
	Level    int              `bson:"level" json:"level"` // For tools - preserves their level
}
