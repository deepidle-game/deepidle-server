package claims

import (
	"deepidle-server/models"
)

type ClaimResult struct {
	CanClaim  bool
	Message   string
	Gained    map[string]int
	NewSlot   bool
	Inventory []models.Item
}

func CalculateResources(char *models.Character, actionCfg models.ActionConfig, timePassed int64) (int, string) {
	if actionCfg.RequiredTool == "" {
		return 0, ""
	}

	toolLevel := 0
	for _, item := range char.Inventory {
		if item.ItemID == actionCfg.RequiredTool {
			toolLevel = item.Level
			break
		}
	}

	if toolLevel == 0 {
		return 0, ""
	}

	cycles := int(timePassed / actionCfg.BaseTimeSec)
	if cycles <= 0 {
		return 0, ""
	}

	totalBaseYield := float64(cycles * actionCfg.BaseAmount)
	multiplier := 1.0 + (float64(toolLevel-1) * actionCfg.BonusPerLevel)
	resourceCount := int(totalBaseYield * multiplier)

	return resourceCount, actionCfg.DropItem
}

func AddResourceToInventory(items []models.Item, itemID string, quantity, maxSlots int) (stacked bool, newSlot bool, newItems []models.Item) {
	for i, item := range items {
		if item.ItemID == itemID {
			items[i].Quantity += quantity
			return true, false, items
		}
	}

	if len(items) >= maxSlots {
		return false, false, items
	}

	items = append(items, models.Item{
		ItemID:   itemID,
		Level:    1,
		Quantity: quantity,
	})
	return false, true, items
}

func ValidateClaim(char *models.Character, actionCfg models.ActionConfig, timePassed int64) (bool, string) {
	if char.CurrentAction == "Idle" || char.ActionStartedAt == 0 {
		return false, "No active action to claim resources from"
	}

	if actionCfg.RequiredTool == "" {
		return false, "Current action does not yield resources"
	}

	toolLevel := 0
	for _, item := range char.Inventory {
		if item.ItemID == actionCfg.RequiredTool {
			toolLevel = item.Level
			break
		}
	}

	if toolLevel == 0 {
		return false, "You don't have the required tool for this action"
	}

	cycles := int(timePassed / actionCfg.BaseTimeSec)
	if cycles <= 0 {
		return false, "Not enough time passed to claim resources"
	}

	return true, ""
}

func GetToolLevel(char *models.Character, toolItemID string) int {
	for _, item := range char.Inventory {
		if item.ItemID == toolItemID {
			return item.Level
		}
	}
	return 0
}

func TimeToCycles(timePassed, baseTimeSec int64) int {
	return int(timePassed / baseTimeSec)
}

func ApplyMultiplier(baseAmount, toolLevel int, bonusPerLevel float64) int {
	multiplier := 1.0 + (float64(toolLevel-1) * bonusPerLevel)
	return int(float64(baseAmount) * multiplier)
}

func NewClaimResult() ClaimResult {
	return ClaimResult{
		Gained:    make(map[string]int),
		Inventory: []models.Item{},
	}
}

func (r *ClaimResult) SetGained(itemID string, quantity int) {
	r.Gained[itemID] = quantity
}

func (r *ClaimResult) SetSuccess(inventory []models.Item, newSlot bool) {
	r.CanClaim = true
	r.NewSlot = newSlot
	r.Inventory = inventory
}

func (r *ClaimResult) SetFailure(message string) {
	r.CanClaim = false
	r.Message = message
}