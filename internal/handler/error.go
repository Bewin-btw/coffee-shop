package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

var ErrUnsupportedContentType = errors.New("unsupported content type")

type ErrorJson struct {
	Error string `json:"error"`
}

func ErrorResponse(w http.ResponseWriter, msg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorResponse := ErrorJson{Error: msg}
	jsonResponse, err := json.Marshal(errorResponse)
	if err != nil {
		http.Error(w, "Failed to generate error response", http.StatusInternalServerError)
		return
	}
	slog.Error(msg)
	w.Write(jsonResponse)
}
