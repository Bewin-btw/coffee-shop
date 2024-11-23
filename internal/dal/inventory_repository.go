package dal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"hot-cofee/internal/config"
	repositories "hot-cofee/internal/dal/utils"
	"hot-cofee/models"
)

type inventoryRepo struct{}

// NewInventoryRepository creates a new instance of InventoryRepository
func NewInventoryRepository() repositories.InventoryRepository {
	return &inventoryRepo{}
}

func (repo *inventoryRepo) ReadInventory() ([]models.InventoryItem, error) {
	var inventory []models.InventoryItem
	StoragePath := config.GetStoragePath()

	file, err := os.OpenFile(filepath.Join(StoragePath, "inventory.json"), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return inventory, errors.New("unable to open inventory: " + err.Error())
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return inventory, errors.New("file info wasn't received: " + err.Error())
	}

	if stat.Size() > 0 {
		if err := json.NewDecoder(file).Decode(&inventory); err != nil {
			return inventory, errors.New("inventory data wasn't received: " + err.Error())
		}
	}
	return inventory, nil
}

func (repo *inventoryRepo) WriteInventory(inventory []models.InventoryItem) error {
	StoragePath := config.GetStoragePath()
	file, err := os.OpenFile(filepath.Join(StoragePath, "inventory.json"), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return errors.New("unable to open inventory: " + err.Error())
	}
	defer file.Close()

	err = file.Truncate(0)
	if err != nil {
		return errors.New("unable to truncate inventory: " + err.Error())
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return errors.New("unable to seek inventory: " + err.Error())
	}

	inventoryData, err := json.MarshalIndent(inventory, "", "    ")
	if err != nil {
		return errors.New("unable to marshal inventory: " + err.Error())
	}
	_, err = file.Write(inventoryData)
	if err != nil {
		return errors.New("unable to write inventory data: " + err.Error())
	}
	return nil
}
