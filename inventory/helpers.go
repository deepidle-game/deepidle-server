package inventory

import (
	"deepidle-server/models"
)

func FindItemIndex(items []models.Item, itemID string) int {
	for i, item := range items {
		if item.ItemID == itemID {
			return i
		}
	}
	return -1
}

func HasMaterials(items []models.Item, required []struct {
	ItemID   string
	Quantity int
}) bool {
	for _, req := range required {
		idx := FindItemIndex(items, req.ItemID)
		if idx == -1 || items[idx].Quantity < req.Quantity {
			return false
		}
	}
	return true
}

func DeductMaterials(items []models.Item, required []struct {
	ItemID   string
	Quantity int
}) []models.Item {
	for _, req := range required {
		idx := FindItemIndex(items, req.ItemID)
		if idx != -1 {
			items[idx].Quantity -= req.Quantity
		}
	}

	newItems := []models.Item{}
	for _, item := range items {
		if item.Quantity > 0 {
			newItems = append(newItems, item)
		}
	}
	return newItems
}

func AddItem(items []models.Item, itemID string, quantity, level, maxSlots int) (bool, []models.Item) {
	for i, item := range items {
		if item.ItemID == itemID {
			items[i].Quantity += quantity
			return true, items
		}
	}

	if len(items) >= maxSlots {
		return false, items
	}

	items = append(items, models.Item{
		ItemID:   itemID,
		Level:    level,
		Quantity: quantity,
	})
	return true, items
}