package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"hot-cofee/internal/service"
	"hot-cofee/models"
)

var OrderService = service.NewOrderService()

func OrderEndpoints(mux *http.ServeMux) {
	mux.HandleFunc("POST /orders", PostOrderHandler)
	mux.HandleFunc("POST /orders/", PostOrderHandler)

	mux.HandleFunc("GET /orders", GetAllOrdersHandler)
	mux.HandleFunc("GET /orders/", GetAllOrdersHandler)

	mux.HandleFunc("GET /orders/{id}", GetOrderByIDHandler)
	mux.HandleFunc("GET /orders/{id}/", GetOrderByIDHandler)

	mux.HandleFunc("PUT /orders/{id}", PutOrderHandler)
	mux.HandleFunc("PUT /orders/{id}/", PutOrderHandler)

	mux.HandleFunc("DELETE /orders/{id}", DeleteOrderByIDHandler)
	mux.HandleFunc("DELETE /orders/{id}/", DeleteOrderByIDHandler)

	mux.HandleFunc("POST /orders/{id}/close", PostOrderCloserHandler)
	mux.HandleFunc("POST /orders/{id}/close/", PostOrderCloserHandler)
}

func GetAllOrdersHandler(w http.ResponseWriter, r *http.Request) {
	orders, err := OrderService.GetAllOrders()
	if err != nil {
		ErrorResponse(w, "Could not retrieve orders data", http.StatusInternalServerError)
		return
	}
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Marshal the items with indentation
	jsonData, err := json.MarshalIndent(orders, "", "    ")
	if err != nil {
		ErrorResponse(w, "Failed to encode inventory items", http.StatusInternalServerError)
		return
	}

	// Write the indented JSON to the response
	if _, err = w.Write(jsonData); err != nil {
		ErrorResponse(w, "Failed to write response", http.StatusInternalServerError)
	}
	slog.Info("Retrieved all orders")
}

func GetOrderByIDHandler(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	ID, err := strconv.Atoi(idString)
	if err != nil {
		ErrorResponse(w, "Invalid order ID", http.StatusBadRequest)
	}
	order, err := OrderService.GetOrderByID(ID)
	if errors.Is(err, service.ErrOrderNotRead) {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonData, err := json.MarshalIndent(order, "", "    ")
	if err != nil {
		ErrorResponse(w, "Failed to encode order", http.StatusInternalServerError)
		return
	}
	// Write the indented JSON to the response
	if _, err = w.Write(jsonData); err != nil {
		ErrorResponse(w, "Failed to write response", http.StatusInternalServerError)
	}
	slog.Info("Retrieved order", "ID", order.ID)
}

func PostOrderHandler(w http.ResponseWriter, r *http.Request) {
	order, err := parseOrder(r)
	if errors.Is(err, ErrUnsupportedContentType) {
		ErrorResponse(w, err.Error(), http.StatusUnsupportedMediaType)
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = OrderService.AddNewOrder(order); errors.Is(err, service.ErrConflict) {
		ErrorResponse(w, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte("Order added successfully")); err != nil {
		ErrorResponse(w, "Failed to write response", http.StatusInternalServerError)
	}
	slog.Info("Added new order", "ID", order.ID)
}

func PostOrderCloserHandler(w http.ResponseWriter, r *http.Request) {
	ordersStruct := service.Order{}
	idString := r.PathValue("id")
	ID, err := strconv.Atoi(idString)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ordersStruct.CloseOrder(ID); errors.Is(err, service.ErrNotExists) {
		ErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte("Order closed successfully")); err != nil {
		ErrorResponse(w, "Failed to write response", http.StatusInternalServerError)
	}
	slog.Info("Closed order", "ID", idString)
}

func PutOrderHandler(w http.ResponseWriter, r *http.Request) {
	order, err := parseOrder(r)
	if errors.Is(err, ErrUnsupportedContentType) {
		ErrorResponse(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	idString := r.PathValue("id")
	ID, err := strconv.Atoi(idString)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = OrderService.ModifyOrder(order, ID); errors.Is(err, service.ErrConflict) {
		ErrorResponse(w, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte("Order is updated successfully")); err != nil {
		ErrorResponse(w, "Failed to write response", http.StatusInternalServerError)
	}
	slog.Info("Updated order", "ID", order.ID)
}

func DeleteOrderByIDHandler(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	ID, err := strconv.Atoi(idString)
	if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = OrderService.DeleteOrder(ID)
	if errors.Is(err, service.ErrOrderNotRead) {
		ErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	slog.Info("Deleted order", "ID", idString)
}

func parseOrder(r *http.Request) (models.Order, error) {
	var order models.Order
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		// Parse JSON payload
		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			return order, fmt.Errorf("invalid JSON payload")
		}
	} else if contentType == "application/x-www-form-urlencoded" {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			return order, fmt.Errorf("invalid form data")
		}
		ID, err := strconv.Atoi(r.FormValue("order_id"))
		if err != nil {
			return order, fmt.Errorf("ID is not an integer")
		}
		var items []models.OrderItem
		itemsJson := r.FormValue("items")
		if err := json.Unmarshal([]byte(itemsJson), &items); err != nil {
			return order, fmt.Errorf("error parsing ingredients: %v", err)
		}

		order = models.Order{
			ID:           ID,
			CustomerName: r.FormValue("customer_name"),
			Items:        items,
			Status:       r.FormValue("status"),
			CreatedAt:    r.FormValue("created_at"),
		}
	} else {
		return order, fmt.Errorf("unsupported content type")
	}

	return order, nil
}
