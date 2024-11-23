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

type menuRepo struct{}

// NewMenuRepository creates a new instance of MenuRepository
func NewMenuRepository() repositories.MenuRepository {
	return &menuRepo{}
}

func (repo *menuRepo) ReadMenu() ([]models.MenuItem, error) {
	var menu []models.MenuItem
	StoragePath := config.GetStoragePath()

	file, err := os.OpenFile(filepath.Join(StoragePath, "menu_items.json"), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return menu, errors.New("unable to open menu file: " + err.Error())
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return menu, errors.New("unable to get file info: " + err.Error())
	}

	if stat.Size() > 0 {
		if err := json.NewDecoder(file).Decode(&menu); err != nil {
			return menu, errors.New("unable to read menu data: " + err.Error())
		}
	}
	return menu, nil
}

func (repo *menuRepo) WriteMenu(menu []models.MenuItem) error {
	StoragePath := config.GetStoragePath()
	file, err := os.OpenFile(filepath.Join(StoragePath, "menu_items.json"), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return errors.New("unable to open menu file: " + err.Error())
	}
	defer file.Close()

	err = file.Truncate(0)
	if err != nil {
		return errors.New("unable to truncate menu file: " + err.Error())
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return errors.New("unable to seek menu file: " + err.Error())
	}

	menuData, err := json.MarshalIndent(menu, "", "    ")
	if err != nil {
		return errors.New("unable to format menu data: " + err.Error())
	}
	_, err = file.Write(menuData)
	if err != nil {
		return errors.New("unable to write menu data: " + err.Error())
	}
	return nil
}
