package repositories

import "hot-cofee/models"

type InventoryRepository interface {
	ReadInventory() ([]models.InventoryItem, error)
	WriteInventory([]models.InventoryItem) error
}

type MenuRepository interface {
	ReadMenu() ([]models.MenuItem, error)
	WriteMenu([]models.MenuItem) error
}

type OrderRepository interface {
	ReadOrder() ([]models.Order, error)
	WriteOrder([]models.Order) error
}
