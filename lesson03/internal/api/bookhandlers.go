package api

import (
	db "bookbackend/internal/database"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateBookRequest struct {
	Title             string `json:"title"`
	Author            string `json:"author"`
	Description       string `json:"description"`
	YearOfPublication int32  `json:"year_of_publication"`
}

type BookHandler struct {
	queries *db.Queries
}

func NewBookHandler(queries *db.Queries) *BookHandler {
	return &BookHandler{queries: queries}
}

func (h *BookHandler) FetchBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.queries.ListBooks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (h *BookHandler) FetchBookByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	book, err := h.queries.GetBook(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req CreateBookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	params := db.CreateBookParams{
		Title:  req.Title,
		Author: req.Author,
	}

	if req.Description != "" {
		params.Description = pgtype.Text{String: req.Description, Valid: true}
	}

	if req.YearOfPublication != 0 {
		params.YearOfPublication = pgtype.Int4{Int32: req.YearOfPublication, Valid: true}
	}

	book, err := h.queries.CreateBook(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id32, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = h.queries.GetBook(r.Context(), int32(id32))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.queries.DeleteBook(r.Context(), int32(id32))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
