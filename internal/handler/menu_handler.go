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

var MenuService = service.NewMenuService()

func MenuEndpoints(mux *http.ServeMux) {
	mux.HandleFunc("POST /menu", PostMenuHandler)
	mux.HandleFunc("POST /menu/", PostMenuHandler)

	mux.HandleFunc("GET /menu", GetAllMenuHandler)
	mux.HandleFunc("GET /menu/", GetAllMenuHandler)

	mux.HandleFunc("GET /menu/{id}", GetMenuByIDHandler)
	mux.HandleFunc("GET /menu/{id}/", GetMenuByIDHandler)

	mux.HandleFunc("PUT /menu/{id}", PutMenuHandler)
	mux.HandleFunc("PUT /menu/{id}/", PutMenuHandler)

	mux.HandleFunc("DELETE /menu/{id}", DeleteMenuByIDHandler)
	mux.HandleFunc("DELETE /menu/{id}/", DeleteMenuByIDHandler)
}

func GetAllMenuHandler(w http.ResponseWriter, r *http.Request) {
	menu, err := MenuService.GetAllMenu()
	if err != nil {
		ErrorResponse(w, "Could not retrieve menu data", http.StatusInternalServerError)
		return
	}
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Marshal the items with indentation
	jsonData, err := json.MarshalIndent(menu, "", "    ")
	if err != nil {
		ErrorResponse(w, "Failed to encode menu items", http.StatusInternalServerError)
		return
	}
	// Write the indented JSON to the response
	if _, err = w.Write(jsonData); err != nil {
		ErrorResponse(w, "Failed to write response", http.StatusInternalServerError)
	}
	slog.Info("Retrieved all menu products")
}

func GetMenuByIDHandler(w http.ResponseWriter, r *http.Request) {
	itemId := r.PathValue("id")
	item, err := MenuService.GetMenuByID(itemId)
	if errors.Is(err, service.ErrMenuNotRead) {
		ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}
	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Marshal the items with indentation
	jsonData, err := json.MarshalIndent(item, "", "    ")
	if err != nil {
		ErrorResponse(w, "Failed to encode menu items", http.StatusInternalServerError)
		return
	}

	// Write the indented JSON to the response
	if _, err = w.Write(jsonData); err != nil {
		ErrorResponse(w, "Failed to write response", http.StatusInternalServerError)
	}
	slog.Info("Retrieved menu item", "ID", item.ID)
}

func DeleteMenuByIDHandler(w http.ResponseWriter, r *http.Request) {
	itemId := r.PathValue("id")
	err := MenuService.DeleteMenuItem(itemId)
	if errors.Is(err, service.ErrMenuNotRead) {
		ErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	slog.Info("Deleted menu item", "ID", itemId)
}

func parseMenuItem(r *http.Request) (models.MenuItem, error) {
	var item models.MenuItem
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		// Parse JSON
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			return item, fmt.Errorf("invalid JSON payload")
		}
	} else if contentType == "application/x-www-form-urlencoded" {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			return item, fmt.Errorf("invalid form data")
		}
		price, err := strconv.ParseFloat(r.FormValue("price"), 64)
		if err != nil {
			return item, fmt.Errorf("price is not a float")
		}

		// Parse ingredients from JSON format within form data
		var ingredients []models.MenuItemIngredient
		ingredientsJSON := r.FormValue("ingredients")
		if err := json.Unmarshal([]byte(ingredientsJSON), &ingredients); err != nil {
			return item, fmt.Errorf("error parsing ingredients: %v", err)
		}

		// Build MenuItem from parsed form values
		item = models.MenuItem{
			ID:          r.FormValue("product_id"),
			Name:        r.FormValue("name"),
			Description: r.FormValue("description"),
			Price:       price,
			Ingredients: ingredients,
		}
	} else {
		return item, ErrUnsupportedContentType
	}

	return item, nil
}

func PostMenuHandler(w http.ResponseWriter, r *http.Request) {
	item, err := parseMenuItem(r)
	if errors.Is(err, ErrUnsupportedContentType) {
		ErrorResponse(w, err.Error(), http.StatusUnsupportedMediaType)
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := MenuService.AddNewMenuItem(item); errors.Is(err, service.ErrConflict) {
		ErrorResponse(w, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte("Menu item added successfully")); err != nil {
		ErrorResponse(w, "Failed to write response", http.StatusInternalServerError)
	}
	slog.Info("Created menu item", "ID", item.ID)
}

func PutMenuHandler(w http.ResponseWriter, r *http.Request) {
	item, err := parseMenuItem(r)
	if errors.Is(err, ErrUnsupportedContentType) {
		ErrorResponse(w, err.Error(), http.StatusUnsupportedMediaType)
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := r.PathValue("id")

	if id != item.ID {
		ErrorResponse(w, "product ID does not match id", http.StatusBadRequest)
		return
	}

	if err := MenuService.ModifyMenuItem(item); errors.Is(err, service.ErrConflict) {
		ErrorResponse(w, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write([]byte("Menu item updated successfully")); err != nil {
		ErrorResponse(w, "Failed to write response", http.StatusInternalServerError)
	}
	slog.Info("Updated menu item", "ID", item.ID)
}
