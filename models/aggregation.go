package models

type TotalSales struct {
	Amount float64 `json:"total_sales"`
}

type PopularItem struct {
	Quantity    int                  `json:"quantity"`
	ID          string               `json:"product_id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Price       float64              `json:"price"`
	Ingredients []MenuItemIngredient `json:"ingredients"`
}
