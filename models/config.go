package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// ActionConfig stores the configuration for a specific action
type ActionConfig struct {
	RequiredTool  string  `bson:"required_tool" json:"required_tool"`
	DropItem      string  `bson:"drop_item" json:"drop_item"`
	BaseTimeSec   int64   `bson:"base_time_sec" json:"base_time_sec"`
	BaseAmount    int     `bson:"base_amount" json:"base_amount"`
	BonusPerLevel float64 `bson:"bonus_per_level" json:"bonus_per_level"` // e.g., 0.20 for 20% bonus per tool level over 1
}

// UpgradeMaterial defines a single material required for upgrade
type UpgradeMaterial struct {
	ItemID   string `bson:"item_id" json:"item_id"`
	Quantity int    `bson:"quantity" json:"quantity"` // Formula: quantity * currentLevel
}

// UpgradeConfig defines materials needed to upgrade an item
type UpgradeConfig struct {
	BaseLevel   int              `bson:"base_level" json:"base_level"` // Starting level for upgrade
	Materials   []UpgradeMaterial `bson:"materials" json:"materials"`
}

// GameConfig is the root configuration object for the server
type GameConfig struct {
	ID        primitive.ObjectID          `bson:"_id,omitempty" json:"id,omitempty"`
	ConfigID  string                     `bson:"config_id" json:"config_id"`
	Actions   map[string]ActionConfig   `bson:"actions" json:"actions"`
	Upgrades  map[string]UpgradeConfig   `bson:"upgrades" json:"upgrades"`
	BaseSlots int                       `bson:"base_slots" json:"base_slots"`
	Resources []string                  `bson:"resources" json:"resources"`
}
