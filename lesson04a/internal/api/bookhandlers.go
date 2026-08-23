package api

import (
	"bookbackend/internal/api/models"
	db "bookbackend/internal/database"
	"log"
	"strings"

	"encoding/json"
	"errors"
	"net/http"

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

func (r CreateBookRequest) Validate() error {
	return validateBook(r.Title, r.Author)
}

type UpdateBookRequest struct {
	Title             string `json:"title"`
	Author            string `json:"author"`
	Description       string `json:"description"`
	YearOfPublication int32  `json:"year_of_publication"`
}

func (r UpdateBookRequest) Validate() error {
	return validateBook(r.Title, r.Author)
}

func validateBook(title, author string) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(author) == "" {
		return errors.New("author is required")
	}
	return nil
}

type BookHandler struct {
	queries db.Querier
}

func NewBookHandler(queries db.Querier) *BookHandler {
	return &BookHandler{queries: queries}
}

func (h *BookHandler) FetchBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.queries.ListBooks(r.Context())
	if err != nil {
		log.Printf("Unable to list books: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	var responses []models.BookResponse
	for _, book := range books {
		responses = append(responses, models.ToBookResponse(book))
	}
	writeJSON(w, http.StatusOK, responses)
}

func (h *BookHandler) FetchBookByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		log.Printf("Invalid book ID: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	book, err := h.queries.GetBook(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("Book not found: %v", err)
			writeError(w, http.StatusNotFound, "Book not found")
			return
		}
		log.Printf("Unable to get book: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, models.ToBookResponse(book))
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req CreateBookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Unable to decode request body: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		log.Printf("Validation error: %v", err)
		writeError(w, http.StatusBadRequest, "Validation error")
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
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("Book created: %v", book)
	writeJSON(w, http.StatusCreated, models.ToBookResponse(book))
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		log.Printf("Invalid book ID: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid book ID")
		return
	}
	_, err = h.queries.DeleteBook(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Book not found")
			return
		}
		log.Printf("Unable to delete book: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusNoContent, nil)
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		log.Printf("Invalid book ID: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid book ID")
		return
	}
	var req UpdateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Unable to decode request body: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		log.Printf("Validation error: %v", err)
		writeError(w, http.StatusBadRequest, "Validation error")
		return
	}

	params := db.UpdateBookParams{
		ID:     id,
		Title:  req.Title,
		Author: req.Author,
	}
	if req.Description != "" {
		params.Description = pgtype.Text{String: req.Description, Valid: true}
	}
	if req.YearOfPublication != 0 {
		params.YearOfPublication = pgtype.Int4{Int32: req.YearOfPublication, Valid: true}
	}
	book, err := h.queries.UpdateBook(r.Context(), params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Book not found")
			return
		}
		log.Printf("Unable to update book: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, models.ToBookResponse(book))
}

func (h *BookHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.FetchBooks)
	r.Get("/{id}", h.FetchBookByID)
	r.Post("/", h.CreateBook)
	r.Put("/{id}", h.UpdateBook)
	r.Delete("/{id}", h.DeleteBook)

	return r
}
