package api

import (
	"bookbackend/internal/api/models"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, models.ErrorResponse{
		Error: message,
	})
}

func parseID(r *http.Request) (int32, error) {
	idStr := chi.URLParam(r, "id")
	value, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, errors.New("id must be a positive integer")
	}
	return int32(value), nil
}
