package service

import (
	"errors"
	"fmt"

	"hot-cofee/internal/dal"
	"hot-cofee/models"
)

type Inventory struct {
	cacheInventory   []models.InventoryItem
	takenIDInventory map[string]int
}

type InventoryService interface {
	LoadInventoryCache() error
	GetAllInventory() ([]models.InventoryItem, error)
	GetInventoryByID(id string) (models.InventoryItem, error)
	AddNewInventoryItem(item models.InventoryItem) error
	DeleteInventoryItem(id string) error
	ModifyInventoryItem(item models.InventoryItem) error
	DeductInventoryItem(ID string, quantity float64) error
}

func NewInventoryService() InventoryService {
	return &Inventory{
		cacheInventory:   []models.InventoryItem{},
		takenIDInventory: make(map[string]int),
	}
}

// LoadInventoryCache loads the inventory data from the file to the cache
func (i *Inventory) LoadInventoryCache() error {
	inventory, err := dal.NewInventoryRepository().ReadInventory()
	if err != nil {
		return errors.Join(ErrInventoryNotRead, err)
	}
	i.cacheInventory = inventory
	i.takenIDInventory = make(map[string]int)
	for j, val := range i.cacheInventory {
		err = validatePostInventory(val)
		if err != nil {
			return errors.Join(ErrConflict, err)
		}
		if _, exists := i.takenIDInventory[val.IngredientID]; exists {
			return ErrConflict
		}
		i.takenIDInventory[val.IngredientID] = j
	}
	return nil
}

// GetAllInventory retrieves all inventory items
func (i *Inventory) GetAllInventory() ([]models.InventoryItem, error) {
	err := i.LoadInventoryCache()
	if err != nil {
		return nil, err
	}
	return i.cacheInventory, nil
}

// GetInventoryByID retrieves a single inventory item by ID
func (i *Inventory) GetInventoryByID(id string) (models.InventoryItem, error) {
	err := i.LoadInventoryCache()
	if err != nil {
		return models.InventoryItem{}, err
	}
	index, exists := i.takenIDInventory[id]
	if !exists || index < 0 || index >= len(i.cacheInventory) {
		return models.InventoryItem{}, fmt.Errorf("item with ingredient ID %s not found", id)
	}
	return i.cacheInventory[index], nil
}

// AddNewInventoryItem adds a new inventory item to the cache and persists it
func (i *Inventory) AddNewInventoryItem(item models.InventoryItem) error {
	err := i.LoadInventoryCache()
	if err != nil {
		return err
	}
	if _, exists := i.takenIDInventory[item.IngredientID]; exists {
		return ErrConflict
	}
	if err := validatePostInventory(item); err != nil {
		return err
	}
	i.cacheInventory = append(i.cacheInventory, item)
	if err := dal.NewInventoryRepository().WriteInventory(i.cacheInventory); err != nil {
		return errors.New("failed to save inventory item")
	}
	return nil
}

// DeleteInventoryItem deletes an inventory item by ID
func (i *Inventory) DeleteInventoryItem(id string) error {
	err := i.LoadInventoryCache()
	if err != nil {
		return err
	}
	index, exists := i.takenIDInventory[id]
	if !exists || index < 0 || index >= len(i.cacheInventory) {
		return fmt.Errorf("item with ingredient ID %s not found", id)
	}
	i.cacheInventory = append(i.cacheInventory[:index], i.cacheInventory[index+1:]...)
	err = dal.NewInventoryRepository().WriteInventory(i.cacheInventory)
	if err != nil {
		return err
	}
	return nil
}

// ModifyInventoryItem modifies an existing inventory item
func (i *Inventory) ModifyInventoryItem(item models.InventoryItem) error {
	err := i.LoadInventoryCache()
	if err != nil {
		return err
	}
	index, exists := i.takenIDInventory[item.IngredientID]
	if !exists || index < 0 || index >= len(i.cacheInventory) {
		return fmt.Errorf("item with ingredient ID %s not found", item.IngredientID)
	}
	if err := validatePostInventory(item); err != nil {
		return err
	}
	if i.cacheInventory[index] == item {
		return ErrNothingToModify
	}
	i.cacheInventory[index] = item
	err = dal.NewInventoryRepository().WriteInventory(i.cacheInventory)
	if err != nil {
		return err
	}
	return nil
}

// DeductInventoryItem deducts a certain quantity from the inventory item
func (i *Inventory) DeductInventoryItem(ID string, quantity float64) error {
	err := i.LoadInventoryCache()
	if err != nil {
		return err
	}
	index, _ := i.takenIDInventory[ID]
	item, err := i.GetInventoryByID(ID)
	if err != nil {
		return err
	}
	i.cacheInventory[index].Quantity = item.Quantity - quantity
	if err := validatePostInventory(i.cacheInventory[index]); err != nil {
		i.cacheInventory[index].Quantity = item.Quantity
		return errors.New("not enough quantity")
	}
	err = dal.NewInventoryRepository().WriteInventory(i.cacheInventory)
	if err != nil {
		return err
	}
	return nil
}
