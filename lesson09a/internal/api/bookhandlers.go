package api

import (
	"bookbackend/internal/api/models"
	db "bookbackend/internal/database"
	"log"
	"strings"
	"time"

	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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

type BorrowBookRequest struct {
	BookID int32 `json:"book_id"`
	UserID int32 `json:"user_id"`
	Days   int32 `json:"days"`
}

type ReturnBookRequest struct {
	BookID int32 `json:"book_id"`
	UserID int32 `json:"user_id"`
}

type BookHandler struct {
	queries QuerierWithTx
	pool    *pgxpool.Pool
}

func NewBookHandler(queries QuerierWithTx, pool *pgxpool.Pool) *BookHandler {
	return &BookHandler{queries: queries, pool: pool}
}

// FetchBooks godoc
// @Summary List all books
// @Description Get a list of all books in the database
// @Tags books
// @Accept json
// @Produce json
// @Success 200 {array} models.BookResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /books [get]
func (h *BookHandler) FetchBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.queries.ListBooks(r.Context())
	if err != nil {
		log.Printf("Unable to list books: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	responses := make([]models.BookResponse, 0, len(books))
	for _, book := range books {
		responses = append(responses, models.ToBookResponse(book))
	}
	writeJSON(w, http.StatusOK, responses)
}

// FetchBookByID godoc
// @Summary Get a book by ID
// @Description Get a single book by its ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Success 200 {object} models.BookResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /books/{id} [get]
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

// CreateBook godoc
// @Summary Create a new book
// @Description Create a new book in the database
// @Tags books
// @Accept json
// @Produce json
// @Param request body CreateBookRequest true "Book details"
// @Success 201 {object} models.BookResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /books [post]
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req CreateBookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Unable to decode request body: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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
		log.Printf("Unable to create book: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	log.Printf("Book created: %d", book.ID)
	writeJSON(w, http.StatusCreated, models.ToBookResponse(book))
}

// DeleteBook godoc
// @Summary Delete a book
// @Description Delete a book by its ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Success 204 "Book deleted successfully"
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /books/{id} [delete]
func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		log.Printf("Invalid book ID: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid book ID")
		return
	}

	_, err = h.queries.DeleteBook(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("Book not found: %v", err)
			writeError(w, http.StatusNotFound, "Book not found")
			return
		}
		log.Printf("Unable to delete book: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateBook godoc
// @Summary Update a book
// @Description Update a book by its ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Param request body UpdateBookRequest true "Book details"
// @Success 200 {object} models.BookResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /books/{id} [put]
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
		writeError(w, http.StatusBadRequest, err.Error())
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

// BorrowBook godoc
// @Summary Borrow a book
// @Description Borrow a book by its ID
// @Tags borrowing
// @Accept json
// @Produce json
// @Param request body BorrowBookRequest true "Borrow details"
// @Success 200 {object} models.BorrowRecordResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /books/borrow [post]
func (h *BookHandler) BorrowBook(w http.ResponseWriter, r *http.Request) {
	var req BorrowBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Unable to decode request body: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Days == 0 {
		req.Days = 14
	}
	tx, err := h.pool.BeginTx(r.Context(), pgx.TxOptions{})
	if err != nil {
		log.Printf("Unable to start transaction: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback(r.Context())

	qtx := h.queries.WithTx(tx)

	//Check if book exists
	book, err := qtx.GetBookForUpdate(r.Context(), req.BookID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Book not found")
			return
		}
		log.Printf("Unable to get book: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if !book.Available {
		// Book is not available for borrowing
		writeError(w, http.StatusConflict, "Book is not available for borrowing")
		return
	}

	_, err = qtx.GetUser(r.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Printf("Unable to get user: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	dueDate := time.Now().Add(time.Duration(req.Days) * 24 * time.Hour)

	borrowRecord, err := qtx.BorrowBook(r.Context(), db.BorrowBookParams{
		BookID:  req.BookID,
		UserID:  req.UserID,
		DueDate: pgtype.Timestamp{Time: dueDate, Valid: true},
	})

	if err != nil {
		log.Printf("Unable to borrow book: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	err = qtx.UpdateBookAvailability(r.Context(), db.UpdateBookAvailabilityParams{
		ID:        req.BookID,
		Available: false,
	})
	if err != nil {
		log.Printf("Unable to update book availability: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if err = tx.Commit(r.Context()); err != nil {
		log.Printf("Unable to commit transaction: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, models.ToBorrowRecordResponse(borrowRecord))
}

// ReturnBook godoc
// @Summary Return a book
// @Description Return a book by its ID
// @Tags borrowing
// @Accept json
// @Produce json
// @Param request body ReturnBookRequest true "Return details"
// @Success 200 {object} models.ReturnRecordResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /books/return [post]
func (h *BookHandler) ReturnBook(w http.ResponseWriter, r *http.Request) {
	var req ReturnBookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Unable to decode request body: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	tx, err := h.pool.BeginTx(r.Context(), pgx.TxOptions{})
	if err != nil {
		log.Printf("Unable to start transaction: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback(r.Context())

	qtx := h.queries.WithTx(tx)
	// Return the book
	returnRecord, err := qtx.ReturnBook(r.Context(), db.ReturnBookParams{
		BookID: req.BookID,
		UserID: req.UserID,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "No active borrow record found")
			return
		}
		log.Printf("Unable to retrieve borrow record: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Mark book as available
	err = qtx.UpdateBookAvailability(r.Context(), db.UpdateBookAvailabilityParams{
		ID:        req.BookID,
		Available: true,
	})

	if err != nil {
		log.Printf("Unable to update book availability: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if err = tx.Commit(r.Context()); err != nil {
		log.Printf("Unable to commit transaction: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, models.ToReturnRecordResponse(returnRecord, "Book returned successfully"))
}

// GetUserBorrowedBooks godoc
// @Summary Get borrowed books for a user
// @Description Get all borrowed books for a specific user by their ID
// @Tags borrowing
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {array} models.BorrowedBooksRowResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /books/user/{id}/borrowed [get]
func (h *BookHandler) GetUserBorrowedBooks(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	books, err := h.queries.GetUserBorrowedBooks(r.Context(), id)
	if err != nil {
		log.Printf("Unable to get user borrowed books: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	var response = make([]models.BorrowedBooksRowResponse, 0, len(books))
	for _, book := range books {
		response = append(response, models.ToBorrowRecordRowResponse(book))
	}
	writeJSON(w, http.StatusOK, response)
}

// GetOverdueBooks godoc
// @Summary Get overdue books
// @Description Get all overdue books
// @Tags borrowing
// @Accept json
// @Produce json
// @Success 200 {array} models.OverdueBooksRowResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /books/overdue [get]
func (h *BookHandler) GetOverdueBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.queries.GetOverdueBooks(r.Context())
	if err != nil {
		log.Printf("Unable to get overdue books: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	var response = make([]models.OverdueBooksRowResponse, 0, len(books))
	for _, book := range books {
		response = append(response, models.ToOverdueBooksRowResponse(book))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *BookHandler) Routes() chi.Router {
	r := chi.NewRouter()
	// Book routes
	r.Get("/", h.FetchBooks)
	r.Get("/{id}", h.FetchBookByID)
	r.Post("/", h.CreateBook)
	r.Put("/{id}", h.UpdateBook)
	r.Delete("/{id}", h.DeleteBook)

	// Borrow and return routes
	r.Post("/borrow", h.BorrowBook)
	r.Post("/return", h.ReturnBook)

	// User borrowed books route
	r.Get("/user/{id}/borrowed", h.GetUserBorrowedBooks)

	// Overdue books route
	r.Get("/overdue", h.GetOverdueBooks)

	return r
}
