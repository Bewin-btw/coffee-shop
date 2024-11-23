package service

import (
	"errors"
	"fmt"
	"time"

	"hot-cofee/internal/dal"
	"hot-cofee/models"
)

type Order struct {
	cacheOrders   []models.Order
	takenIDOrders map[int]int
}

type OrderService interface {
	GetAllOrders() ([]models.Order, error)
	GetOrderByID(ID int) (models.Order, error)
	AddNewOrder(order models.Order) error
	CloseOrder(ID int) error
	DeleteOrder(ID int) error
	ModifyOrder(order models.Order, ID int) error
	LoadOrdersCache() error
}

func NewOrderService() OrderService {
	return &Order{
		cacheOrders:   []models.Order{},
		takenIDOrders: make(map[int]int),
	}
}

func (o *Order) findOrderIndexByID(ID int) (int, error) {
	index, exists := o.takenIDOrders[ID]
	if !exists || index < 0 || index >= len(o.cacheOrders) {
		return -1, fmt.Errorf("order with ID %d not found", ID)
	}
	return index, nil
}

func (o *Order) LoadOrdersCache() error {
	orders, err := dal.NewOrderRepository().ReadOrder()
	if err != nil {
		return errors.Join(ErrOrderNotRead, err)
	}
	o.cacheOrders = orders
	o.takenIDOrders = make(map[int]int)
	if err = validateOrders(orders); err != nil {
		return err
	}
	for i, val := range o.cacheOrders {
		o.takenIDOrders[val.ID] = i
	}
	return nil
}

func (o *Order) GetAllOrders() ([]models.Order, error) {
	err := o.LoadOrdersCache()
	if err != nil {
		return nil, err
	}
	if len(o.cacheOrders) == 0 {
		return o.cacheOrders, errors.New("no orders in orders in orders storage")
	}
	return o.cacheOrders, nil
}

func (o *Order) GetOrderByID(id int) (models.Order, error) {
	err := o.LoadOrdersCache()
	if err != nil {
		return models.Order{}, err
	}
	index, err := o.findOrderIndexByID(id)
	if err != nil {
		return models.Order{}, err
	}
	return o.cacheOrders[index], nil
}

func (o *Order) AddNewOrder(order models.Order) error {
	err := o.LoadOrdersCache()
	if err != nil {
		return err
	}
	if len(o.cacheOrders) == 0 {
		order.ID = 0
	} else {
		lastId := o.cacheOrders[len(o.cacheOrders)-1].ID
		order.ID = lastId + 1
	}
	if err := validateOrder(order); err != nil {
		return err
	}
	order.Status = "Open"
	order.CreatedAt = time.Now().Format(time.DateTime)
	if _, exists := o.takenIDOrders[order.ID]; exists {
		return ErrConflict
	}
	o.cacheOrders = append(o.cacheOrders, order)
	if err := dal.NewOrderRepository().WriteOrder(o.cacheOrders); err != nil {
		return err
	}
	return nil
}

func (o *Order) CloseOrder(ID int) error {
	m := NewMenuService()
	// Load orders from cache
	order, err := o.GetOrderByID(ID)
	if err != nil {
		return err
	}
	err = o.LoadOrdersCache()
	if err != nil {
		return err
	}
	if err := validateOrder(order); err != nil {
		return err
	}
	if err := validateCloseOrder(order); err != nil {
		return err
	}
	for _, product := range order.Items {
		if err := validateDeductCheckIngredients(product.ProductID, float64(product.Quantity)); err != nil {
			return err
		}
		if err := m.DeductMenuProduct(product.ProductID, float64(product.Quantity)); err != nil {
			return err
		}
	}
	order.Status = "Closed"
	o.cacheOrders[ID] = order
	if err := dal.NewOrderRepository().WriteOrder(o.cacheOrders); err != nil {
		return err
	}
	return nil
}

func (o *Order) DeleteOrder(ID int) error {
	err := o.LoadOrdersCache()
	if err != nil {
		return err
	}
	index, exists := o.takenIDOrders[ID]
	if !exists || index < 0 || index >= len(o.cacheOrders) {
		return fmt.Errorf("order with id  %d not found", ID)
	}
	o.cacheOrders = append(o.cacheOrders[:index], o.cacheOrders[index+1:]...)
	err = dal.NewOrderRepository().WriteOrder(o.cacheOrders)
	if err != nil {
		return err
	}
	return nil
}

func (o *Order) ModifyOrder(order models.Order, ID int) error {
	err := o.LoadOrdersCache()
	if err != nil {
		return err
	}
	index, exists := o.takenIDOrders[ID]
	if !exists || index < 0 || index >= len(o.cacheOrders) {
		return fmt.Errorf("order with id  %d not found", order.ID)
	}
	order = orderInit(order, o.cacheOrders[index])
	if err := validateModifying(order, o.cacheOrders[index]); err != nil {
		return err
	}
	if err = validateOrder(order); err != nil {
		return err
	}
	o.cacheOrders[index] = order
	if err := dal.NewOrderRepository().WriteOrder(o.cacheOrders); err != nil {
		return errors.New("failed to modify order")
	}
	return nil
}

func orderInit(modifiedOrder, originalOrder models.Order) models.Order {
	if modifiedOrder.CreatedAt == "" {
		modifiedOrder.CreatedAt = originalOrder.CreatedAt
	}
	if modifiedOrder.CustomerName == "" {
		modifiedOrder.CustomerName = originalOrder.CustomerName
	}
	if modifiedOrder.Items == nil {
		modifiedOrder.Items = originalOrder.Items
	}
	if modifiedOrder.Status == "" {
		modifiedOrder.Status = originalOrder.Status
	}
	return modifiedOrder
}
