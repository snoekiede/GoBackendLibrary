package api

import (
	db "bookbackend/internal/database"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

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

type BorrowBookRequest struct {
	BookID int32 `json:"book_id"`
	UserID int32 `json:"user_id"`
	Days   int32 `json:"days"` // How many days to borrow for
}

type ReturnBookRequest struct {
	BookID int32 `json:"book_id"`
	UserID int32 `json:"user_id"`
}

type BookStore struct {
	db   QuerierWithTx
	pool *pgxpool.Pool
}

func NewBookStore(db QuerierWithTx, pool *pgxpool.Pool) *BookStore {
	return &BookStore{db: db, pool: pool}
}

// CreateBook godoc
// @Summary Create a new book
// @Description Create a new book in the database
// @Tags books
// @Accept json
// @Produce json
// @Param request body CreateBookRequest true "Book details"
// @Success 201 {object} db.Book
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /books [post]
func (store *BookStore) CreateBook(w http.ResponseWriter, r *http.Request) {
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

	book, err := store.db.CreateBook(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

// FetchBooks godoc
// @Summary List all books
// @Description Get all books from the database
// @Tags books
// @Accept json
// @Produce json
// @Success 200 {array} db.Book
// @Failure 500 {object} map[string]string
// @Router /books [get]
func (store *BookStore) FetchBooks(w http.ResponseWriter, r *http.Request) {
	books, err := store.db.ListBooks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// FetchBookByID godoc
// @Summary Get a book by ID
// @Description Get a single book by its ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Success 200 {object} db.Book
// @Failure 404 {object} map[string]string
// @Router /books/{id} [get]
func (store *BookStore) FetchBookByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	book, err := store.db.GetBook(r.Context(), int32(id))
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

// DeleteBook godoc
// @Summary Delete a book
// @Description Delete a book by its ID
// @Tags books
// @Accept json
// @Produce json
// @Param id path int true "Book ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /books/{id} [delete]
func (store *BookStore) DeleteBook(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id32, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = store.db.GetBook(r.Context(), int32(id32))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = store.db.DeleteBook(r.Context(), int32(id32))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// the borrowing logic

// BorrowBook godoc
// @Summary Borrow a book
// @Description Borrow a book for a specified number of days
// @Tags borrowing
// @Accept json
// @Produce json
// @Param request body BorrowBookRequest true "Borrow details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /borrow [post]
func (store *BookStore) BorrowBook(w http.ResponseWriter, r *http.Request) {
	var req BorrowBookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Days == 0 {
		req.Days = 14
	}

	tx, err := store.pool.BeginTx(r.Context(), pgx.TxOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer tx.Rollback(r.Context())

	qtx := store.db.WithTx(tx)

	book, err := qtx.GetBookForUpdate(r.Context(), req.BookID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Book not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !book.Available.Bool {
		http.Error(w, "Book not available", http.StatusServiceUnavailable)
		return
	}

	_, err = qtx.GetUser(r.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dueDate := time.Now().Add(time.Duration(req.Days) * 24 * time.Hour)

	borrowRecord, err := qtx.BorrowBook(r.Context(), db.BorrowBookParams{
		BookID: req.BookID,
		UserID: req.UserID,
		DueDate: pgtype.Timestamp{
			Time:  dueDate,
			Valid: true,
		},
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = qtx.UpdateBookAvailability(r.Context(), db.UpdateBookAvailabilityParams{
		ID:        req.BookID,
		Available: pgtype.Bool{Bool: false, Valid: true},
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"borrow_record": borrowRecord,
		"book":          book,
		"message":       "Book borrowed successfully",
	})

}

// ReturnBook godoc
// @Summary Return a borrowed book
// @Description Return a book that was previously borrowed
// @Tags borrowing
// @Accept json
// @Produce json
// @Param request body ReturnBookRequest true "Return details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /return [post]
func (store *BookStore) ReturnBook(w http.ResponseWriter, r *http.Request) {
	var req ReturnBookRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := store.pool.BeginTx(r.Context(), pgx.TxOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	qtx := store.db.WithTx(tx)

	// Return the book
	returnRecord, err := qtx.ReturnBook(r.Context(), db.ReturnBookParams{
		BookID: req.BookID,
		UserID: req.UserID,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "No active borrow record found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Mark book as available
	err = qtx.UpdateBookAvailability(r.Context(), db.UpdateBookAvailabilityParams{
		ID:        req.BookID,
		Available: pgtype.Bool{Bool: true, Valid: true},
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"return_record": returnRecord,
		"message":       "Book returned successfully",
	})
}

// GetUserBorrowedBooks godoc
// @Summary Get user's borrowed books
// @Description Get all books currently borrowed by a specific user
// @Tags borrowing
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {array} db.GetUserBorrowedBooksRow
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id}/borrowed [get]
func (store *BookStore) GetUserBorrowedBooks(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	books, err := store.db.GetUserBorrowedBooks(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// GetOverdueBooks godoc
// @Summary Get overdue books
// @Description Get all books that are overdue for return
// @Tags borrowing
// @Accept json
// @Produce json
// @Success 200 {array} db.GetOverdueBooksRow
// @Failure 500 {object} map[string]string
// @Router /overdue [get]
func (store *BookStore) GetOverdueBooks(w http.ResponseWriter, r *http.Request) {
	books, err := store.db.GetOverdueBooks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}
