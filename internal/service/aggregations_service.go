package service

import (
	"errors"
	"sort"

	"hot-cofee/models"
)

var (
	ErrNotFoundID             = errors.New("id was not found")
	ErrUnsupportedContentType = errors.New("unsupported content type")
)

func GetTotalSales() (models.TotalSales, error) {
	m := NewMenuService()
	totalSales := models.TotalSales{}
	ordersStruct := Order{}

	err := ordersStruct.LoadOrdersCache()
	if err != nil {
		return totalSales, err
	}
	err = m.LoadMenuCache()
	if err != nil {
		return totalSales, err
	}
	if len(ordersStruct.cacheOrders) == 0 {
		return totalSales, ErrOrderNotRead
	}
	for _, order := range ordersStruct.cacheOrders {
		if order.Status == "Closed" {
			for _, product := range order.Items {
				if err = validateAggregation(product); err != nil {
					return totalSales, err
				}
				menu, errMenu := m.GetMenuByID(product.ProductID)
				if errMenu != nil {
					return models.TotalSales{}, err
				}
				totalSales.Amount = totalSales.Amount + float64(product.Quantity)*menu.Price
			}
		} else if order.Status != "Open" && order.Status != "Closed" {
			return models.TotalSales{}, errors.New("order is not closed")
		}
	}
	return totalSales, nil
}

func GetPopularItems() ([]models.PopularItem, error) {
	o := NewOrderService()

	allOrders, err := o.GetAllOrders()
	if err != nil {
		return []models.PopularItem{}, err
	}
	if len(allOrders) == 0 {
		return []models.PopularItem{}, ErrOrderNotRead
	}
	SumProdID := map[string]int{}

	for _, order := range allOrders {
		if order.Status == "closed" {
			for _, product := range order.Items {
				if err = validateAggregation(product); err != nil {
					return []models.PopularItem{}, err
				}
				if product.Quantity <= 0 {
					return []models.PopularItem{}, errors.New("quantity is <= 0")
				}
				SumProdID[product.ProductID] = SumProdID[product.ProductID] + product.Quantity
			}
		} else if order.Status != "Open" && order.Status != "Closed" {
			return []models.PopularItem{}, errors.New("order is neither closed nor open")
		}
	}
	return GetTopItemsByQuantity(SumProdID, 3), nil
}

// Helper function to get top N items by quantity
func GetTopItemsByQuantity(productQuantities map[string]int, topN int) []models.PopularItem {
	m := NewMenuService()
	var quantities []models.OrderItem
	for id, quantity := range productQuantities {
		quantities = append(quantities, models.OrderItem{id, quantity})
	}

	// Sort products by quantity in descending order
	sort.Slice(quantities, func(i, j int) bool {
		return quantities[i].Quantity > quantities[j].Quantity
	})

	// Get the top N items (or fewer if there are less than N products)
	var topItems []models.PopularItem
	for i := 0; i < len(quantities) && i < topN; i++ {
		menu, menuErr := m.GetMenuByID(quantities[i].ProductID)
		if menuErr != nil {
			return []models.PopularItem{}
		}
		topItems = append(topItems, models.PopularItem{
			Quantity:    quantities[i].Quantity,
			ID:          menu.ID,
			Name:        menu.Name,
			Description: menu.Description,
			Price:       menu.Price,
			Ingredients: menu.Ingredients,
		})
	}

	return topItems
}
