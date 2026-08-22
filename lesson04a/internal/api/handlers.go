package api

import (
	"bookbackend/internal/api/models"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
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
	value, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return int32(value), nil
}
