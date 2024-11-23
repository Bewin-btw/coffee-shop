package service

import (
	"errors"
	"fmt"

	"hot-cofee/models"
)

var (
	ErrNotExists        = errors.New("resource not found")
	ErrIDNotExist       = errors.New("item with this id does not exists")
	ErrZeroLengthID     = errors.New("item cant have 0 length id")
	ErrConflict         = errors.New("item with this ID already exists")
	ErrInventoryNotRead = errors.New("inventory was not read")
	ErrMenuNotRead      = errors.New("menu was not read")
	ErrOrderNotRead     = errors.New("orders were not read")
	ErrNothingToModify  = errors.New("nothing to modify")
	ErrMalformedContent = errors.New("malformed content")
	ErrNotFound         = errors.New("not found")
)

func validatePostInventory(item models.InventoryItem) error {
	if item.IngredientID == "" {
		return errors.New("ingredient ID cannot be empty")
	} else if item.Quantity < 0 {
		return errors.New("quantity cannot be negative")
	} else if item.Unit == "" {
		return errors.New("unit cannot be empty")
	} else if item.Name == "" {
		return errors.New("name cannot be empty")
	}

	return nil
}

func validatePostMenu(item models.MenuItem) error {
	if item.ID == "" {
		return errors.New("product ID cannot be empty")
	} else if item.Price <= 0 {
		return errors.New("price cannot be negative or zero")
	} else if item.Description == "" {
		return errors.New("description cannot be empty")
	} else if item.Name == "" {
		return errors.New("name cannot be empty")
	} else if len(item.Ingredients) < 1 {
		return errors.New("number of ingredients cannot be less than 1")
	}
	return nil
}

func validatePostMenuIngredients(Ingredients []models.MenuItemIngredient) error {
	takenIDMenuInventory := make(map[string]int)
	for j, val := range Ingredients {
		if _, exists := takenIDMenuInventory[val.IngredientID]; exists {
			return errors.New("duplicated ingredient ID")
		}
		takenIDMenuInventory[val.IngredientID] = j

		if val.Quantity < 0 {
			return fmt.Errorf("item with quantity %v is less than 0", val.Quantity)
		}
	}
	return nil
}

func equalSlices[T comparable](slice1, slice2 []T) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}

func validateOrders(Orders []models.Order) error {
	takenIdOrder := make(map[int]int)
	for i, val := range Orders {
		if _, exists := takenIdOrder[val.ID]; exists {
			return errors.New("duplicated order id")
		}
		takenIdOrder[val.ID] = i
		if _, exists := takenIdOrder[val.ID]; !exists {
			return fmt.Errorf("item with order ID %d does not exists", val.ID)
		}
		for _, items := range val.Items {
			if items.Quantity < 1 {
				return fmt.Errorf("item with quantity %v is less than 1", items.Quantity)
			}
		}
	}
	return nil
}

func validateOrder(order models.Order) error {
	varTakenIdOrder := make(map[string]int)
	m := NewMenuService()
	if order.ID < 0 {
		return errors.New("order ID cannot be negative")
	} else if len(order.Items) == 0 {
		return errors.New("empty order")
	} else if order.CustomerName == "" {
		return errors.New("customer name cannot be empty")
	} else if order.Items == nil {
		return errors.New("empty order")
	}
	for i, item := range order.Items {
		product, err := m.GetMenuByID(item.ProductID)
		if err != nil {
			return err
		}
		if _, exists := varTakenIdOrder[item.ProductID]; exists {
			return errors.New("duplicated products in order")
		}
		varTakenIdOrder[item.ProductID] = i
		if item.Quantity <= 0 {
			return fmt.Errorf("item with quantity %v is less than or equal to 0", item.Quantity)
		}
		if err := validatePostMenu(product); err != nil {
			return err
		}
	}
	return nil
}

func validateCloseOrder(order models.Order) error {
	if order.ID < 0 {
		return errors.New("order ID cannot be negative")
	}
	if order.CustomerName == "" {
		return errors.New("customer name cannot be empty")
	}
	if order.Items == nil {
		return errors.New("items cannot be null")
	}
	if order.Status == "Closed" {
		return errors.New("order is already closed")
	}
	return nil
}

func validateDeductCheckIngredients(productID string, quantity float64) error {
	m := NewMenuService()

	item, err := m.GetMenuByID(productID)
	if err != nil {
		return err
	}
	for _, ingredient := range item.Ingredients {
		requiredQuantity := ingredient.Quantity * quantity
		if err := CheckInventoryAvailability(ingredient.IngredientID, requiredQuantity); err != nil {
			return fmt.Errorf("not enough %s (required: %.2f)", ingredient.IngredientID, requiredQuantity)
		}

	}
	return nil
}

func CheckInventoryAvailability(ingredientID string, requiredQuantity float64) error {
	i := NewInventoryService()
	// Get the inventory item by its ID
	item, err := i.GetInventoryByID(ingredientID)
	if err != nil {
		return err
	}
	// Check if there is enough quantity
	if item.Quantity < requiredQuantity {
		return fmt.Errorf("not enough quantity for ingredient %s", ingredientID)
	}
	return nil
}

func validateModifying(modifiedOrder, originalOrder models.Order) error {
	if modifiedOrder.ID != originalOrder.ID {
		return errors.New("order with id does not match")
	}
	if originalOrder.Status != modifiedOrder.Status {
		return errors.New("modifying status is not permitted")
	}
	if modifiedOrder.Status != "Open" && modifiedOrder.Status != "Closed" {
		return errors.New("wrong order status (should be \"Closed\" or \"Open\")")
	}
	if originalOrder.CreatedAt != modifiedOrder.CreatedAt {
		return errors.New("modifying created time is not permitted")
	}
	if originalOrder.ID == modifiedOrder.ID &&
		originalOrder.CustomerName == modifiedOrder.CustomerName &&
		equalSlices(originalOrder.Items, modifiedOrder.Items) &&
		originalOrder.Status == modifiedOrder.Status &&
		originalOrder.CreatedAt == modifiedOrder.CreatedAt {

		return ErrNothingToModify
	}
	return nil
}

func validateAggregation(product models.OrderItem) error {
	m := NewMenuService()

	item, err := m.GetMenuByID(product.ProductID)
	if err != nil {
		return ErrNotFoundID
	}

	if item.Price <= 0 {
		return errors.New("price is <= 0")
	}
	if product.Quantity <= 0 {
		return errors.New("quantity is <= 0")
	}
	return nil
}
