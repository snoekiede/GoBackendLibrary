package api

import (
	"bookbackend/internal/api/models"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

const (
	PAGE_PARAM_NAME = "page"
	SIZE_PARAM_NAME = "size"
	MAX_PAGE_SIZE   = 100
	DEFAULT_PAGE    = 1
	DEFAULT_SIZE    = 10
	MAX_BODY_BYTES  = 1 << 20 // 1 MiB
)

// decodeJSON decodes exactly one JSON value from the request body into dst.
// It enforces a size limit, rejects unknown fields and trailing data, and on
// failure writes a client-facing error response and returns false.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, MAX_BODY_BYTES)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &maxErr):
			writeError(w, http.StatusRequestEntityTooLarge, "Request body too large")
		case errors.Is(err, io.EOF):
			writeError(w, http.StatusBadRequest, "Request body must not be empty")
		case errors.As(err, &syntaxErr), errors.Is(err, io.ErrUnexpectedEOF):
			writeError(w, http.StatusBadRequest, "Request body contains malformed JSON")
		case errors.As(err, &typeErr):
			writeError(w, http.StatusBadRequest, "Request body contains an invalid value for field \""+typeErr.Field+"\"")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			writeError(w, http.StatusBadRequest, "Request body contains an unknown field")
		default:
			writeError(w, http.StatusBadRequest, "Invalid request body")
		}
		return false
	}

	// A second Decode must hit EOF; anything else is trailing data.
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "Request body too large")
		} else {
			writeError(w, http.StatusBadRequest, "Request body must contain a single JSON value")
		}
		return false
	}
	return true
}

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
