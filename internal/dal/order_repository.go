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

type orderRepo struct{}

// NewOrderRepository creates a new instance of OrderRepository
func NewOrderRepository() repositories.OrderRepository {
	return &orderRepo{}
}

func (repo *orderRepo) ReadOrder() ([]models.Order, error) {
	var orders []models.Order
	StoragePath := config.GetStoragePath()

	file, err := os.OpenFile(filepath.Join(StoragePath, "orders.json"), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return orders, errors.New("unable to open order file: " + err.Error())
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return orders, errors.New("unable to get file info: " + err.Error())
	}

	if stat.Size() > 0 {
		if err := json.NewDecoder(file).Decode(&orders); err != nil {
			return orders, errors.New("unable to read order data: " + err.Error())
		}
	}
	return orders, nil
}

func (repo *orderRepo) WriteOrder(orders []models.Order) error {
	StoragePath := config.GetStoragePath()
	file, err := os.OpenFile(filepath.Join(StoragePath, "orders.json"), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return errors.New("unable to open order file: " + err.Error())
	}
	defer file.Close()

	err = file.Truncate(0)
	if err != nil {
		return errors.New("unable to truncate order file: " + err.Error())
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return errors.New("unable to seek order file: " + err.Error())
	}

	orderData, err := json.MarshalIndent(orders, "", "    ")
	if err != nil {
		return errors.New("unable to format order data: " + err.Error())
	}
	_, err = file.Write(orderData)
	if err != nil {
		return errors.New("unable to write order data: " + err.Error())
	}
	return nil
}
