package service

import (
	"errors"
	"fmt"

	"hot-cofee/internal/dal"
	"hot-cofee/models"
)

type Menu struct {
	cacheMenu   []models.MenuItem
	takenIDMenu map[string]int
}

type MenuService interface {
	LoadMenuCache() error
	GetAllMenu() ([]models.MenuItem, error)
	GetMenuByID(id string) (models.MenuItem, error)
	DeleteMenuItem(id string) error
	AddNewMenuItem(item models.MenuItem) error
	ModifyMenuItem(item models.MenuItem) error
	DeductMenuProduct(ID string, quantity float64) error
}

func NewMenuService() MenuService {
	return &Menu{
		cacheMenu:   []models.MenuItem{},
		takenIDMenu: make(map[string]int),
	}
}

func (m *Menu) LoadMenuCache() error {
	menu, err := dal.NewMenuRepository().ReadMenu()
	if err != nil {
		return errors.Join(ErrInventoryNotRead, err)
	}
	m.cacheMenu = menu
	m.takenIDMenu = make(map[string]int)

	for i, val := range m.cacheMenu {
		if _, exists := m.takenIDMenu[val.ID]; exists {
			return ErrConflict
		}

		err = validatePostMenu(val)
		if err != nil {
			return errors.Join(ErrConflict, err)
		}
		err = validatePostMenuIngredients(val.Ingredients)
		if err != nil {
			return errors.Join(ErrConflict, err)
		}
		m.takenIDMenu[val.ID] = i
	}
	return nil
}

func (m *Menu) GetAllMenu() ([]models.MenuItem, error) {
	err := m.LoadMenuCache()
	if err != nil {
		return nil, err
	}
	return m.cacheMenu, nil
}

func (m *Menu) GetMenuByID(id string) (models.MenuItem, error) {
	err := m.LoadMenuCache()
	if err != nil {
		return models.MenuItem{}, err
	}
	index, exists := m.takenIDMenu[id]
	// Check if the item exists in the cacheInventory using the id index
	if !exists || index < 0 || index >= len(m.cacheMenu) {
		return models.MenuItem{}, fmt.Errorf("item with product ID %s not found", id)
	}

	return m.cacheMenu[index], nil
}

func (m *Menu) DeleteMenuItem(id string) error {
	err := m.LoadMenuCache()
	if err != nil {
		return err
	}
	index, exists := m.takenIDMenu[id]
	// Check if the item exists in the cacheInventory using the id index
	if !exists || index < 0 || index >= len(m.cacheMenu) {
		return fmt.Errorf("item with product ID %s not found", id)
	}
	// Remove the item from the cacheInventory slice
	m.cacheMenu = append(m.cacheMenu[:index], m.cacheMenu[index+1:]...)

	err = dal.NewMenuRepository().WriteMenu(m.cacheMenu)
	if err != nil {
		return err
	}
	return nil
}

func (m *Menu) AddNewMenuItem(item models.MenuItem) error {
	err := m.LoadMenuCache()
	if err != nil {
		return err
	}
	// Check if the ID is already taken
	if _, exists := m.takenIDMenu[item.ID]; exists {
		return ErrConflict
	}
	// Validate item
	if err = validatePostMenu(item); err != nil {
		return err
	}
	err = validatePostMenuIngredients(item.Ingredients)
	if err != nil {
		return err
	}

	m.cacheMenu = append(m.cacheMenu, item)
	if err := dal.NewMenuRepository().WriteMenu(m.cacheMenu); err != nil {
		return errors.New("failed to save menu item")
	}

	return nil
}

func (m *Menu) ModifyMenuItem(item models.MenuItem) error {
	err := m.LoadMenuCache()
	if err != nil {
		return err
	}
	index, exists := m.takenIDMenu[item.ID]
	// Check if the ID is already taken
	if !exists || index < 0 || index >= len(m.cacheMenu) {
		return fmt.Errorf("item with product ID %s not found", item.ID)
	}
	// Validate item
	if err = validatePostMenu(item); err != nil {
		return err
	}
	err = validatePostMenuIngredients(item.Ingredients)
	if err != nil {
		return err
	}

	if m.cacheMenu[index].Description == item.Description &&
		m.cacheMenu[index].ID == item.ID &&
		m.cacheMenu[index].Name == item.Name &&
		m.cacheMenu[index].Price == item.Price &&
		equalSlices(m.cacheMenu[index].Ingredients, item.Ingredients) {
		return ErrNothingToModify
	}

	m.cacheMenu[index] = item
	if err := dal.NewMenuRepository().WriteMenu(m.cacheMenu); err != nil {
		return errors.New("failed to modify menu item")
	}

	return nil
}

func (m *Menu) DeductMenuProduct(ID string, quantity float64) error {
	i := NewInventoryService()
	err := m.LoadMenuCache()
	if err != nil {
		return err
	}
	item, err := m.GetMenuByID(ID)
	if err != nil {
		return err
	}
	index, exists := m.takenIDMenu[item.ID]
	if !exists || index < 0 || index >= len(m.cacheMenu) {
		return fmt.Errorf("item with product ID %s not found", item.ID)
	}
	if err = validatePostMenu(item); err != nil {
		return err
	}
	err = validatePostMenuIngredients(item.Ingredients)
	if err != nil {
		return err
	}
	for _, ingredient := range item.Ingredients {
		if err := i.DeductInventoryItem(ingredient.IngredientID, ingredient.Quantity*quantity); err != nil {
			return err
		}
	}
	if err := dal.NewMenuRepository().WriteMenu(m.cacheMenu); err != nil {
		return errors.New("failed to modify menu item")
	}
	return nil
}
