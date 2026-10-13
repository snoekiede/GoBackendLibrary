package api

import (
	"bookbackend/internal/api/models"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

const (
	PAGE_PARAM_NAME = "page"
	SIZE_PARAM_NAME = "size"
	MAX_PAGE_SIZE   = 100
	DEFAULT_PAGE    = 1
	DEFAULT_SIZE    = 10
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

func parsePageAndSize(r *http.Request) (int32, int32) {
	page := int32(DEFAULT_PAGE)
	size := int32(DEFAULT_SIZE)

	if pStr := r.URL.Query().Get(PAGE_PARAM_NAME); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = int32(p)
		}
	}

	if sStr := r.URL.Query().Get(SIZE_PARAM_NAME); sStr != "" {
		if s, err := strconv.Atoi(sStr); err == nil && s > 0 {
			if s > MAX_PAGE_SIZE {
				s = MAX_PAGE_SIZE
			}
			size = int32(s)
		}
	}

	return page, size
}
