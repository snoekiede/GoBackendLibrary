package api

import (
	db "bookbackend/internal/database"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type MockQueries struct {
	db.Querier
	CreateBookFunc             func(ctx context.Context, params db.CreateBookParams) (db.Book, error)
	ListBooksFunc              func(ctx context.Context) ([]db.Book, error)
	GetBookFunc                func(ctx context.Context, id int32) (db.Book, error)
	DeleteBookFunc             func(ctx context.Context, id int32) error
	GetBookForUpdateFunc       func(ctx context.Context, id int32) (db.Book, error)
	GetUserFunc                func(ctx context.Context, id int32) (db.User, error)
	BorrowBookFunc             func(ctx context.Context, params db.BorrowBookParams) (db.BorrowedBook, error)
	UpdateBookAvailabilityFunc func(ctx context.Context, params db.UpdateBookAvailabilityParams) error
	ReturnBookFunc             func(ctx context.Context, params db.ReturnBookParams) (db.BorrowedBook, error)
	GetUserBorrowedBooksFunc   func(ctx context.Context, userID int32) ([]db.GetUserBorrowedBooksRow, error)
	GetOverdueBooksFunc        func(ctx context.Context) ([]db.GetOverdueBooksRow, error)
}

func (m *MockQueries) GetBookForUpdate(ctx context.Context, id int32) (db.Book, error) {
	if m.GetBookForUpdateFunc != nil {
		return m.GetBookForUpdateFunc(ctx, id)
	}
	return db.Book{}, nil
}

func (m *MockQueries) GetUser(ctx context.Context, id int32) (db.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, id)
	}
	return db.User{}, nil
}

func (m *MockQueries) BorrowBook(ctx context.Context, params db.BorrowBookParams) (db.BorrowedBook, error) {
	if m.BorrowBookFunc != nil {
		return m.BorrowBookFunc(ctx, params)
	}
	return db.BorrowedBook{}, nil
}

func (m *MockQueries) UpdateBookAvailability(ctx context.Context, params db.UpdateBookAvailabilityParams) error {
	if m.UpdateBookAvailabilityFunc != nil {
		return m.UpdateBookAvailabilityFunc(ctx, params)
	}
	return nil
}

func (m *MockQueries) ReturnBook(ctx context.Context, params db.ReturnBookParams) (db.BorrowedBook, error) {
	if m.ReturnBookFunc != nil {
		return m.ReturnBookFunc(ctx, params)
	}
	return db.BorrowedBook{}, nil
}

func (m *MockQueries) GetUserBorrowedBooks(ctx context.Context, userID int32) ([]db.GetUserBorrowedBooksRow, error) {
	if m.GetUserBorrowedBooksFunc != nil {
		return m.GetUserBorrowedBooksFunc(ctx, userID)
	}
	return []db.GetUserBorrowedBooksRow{}, nil
}

func (m *MockQueries) GetOverdueBooks(ctx context.Context) ([]db.GetOverdueBooksRow, error) {
	if m.GetOverdueBooksFunc != nil {
		return m.GetOverdueBooksFunc(ctx)
	}
	return []db.GetOverdueBooksRow{}, nil
}

func (m *MockQueries) WithTx(tx pgx.Tx) *db.Queries {
	return db.New(tx)
}

func (m *MockQueries) CreateBook(ctx context.Context, params db.CreateBookParams) (db.Book, error) {
	if m.CreateBookFunc != nil {
		return m.CreateBookFunc(ctx, params)
	}
	return db.Book{}, nil
}

func (m *MockQueries) ListBooks(ctx context.Context) ([]db.Book, error) {
	if m.ListBooksFunc != nil {
		return m.ListBooksFunc(ctx)
	}
	return []db.Book{}, nil
}

func (m *MockQueries) GetBook(ctx context.Context, id int32) (db.Book, error) {
	if m.GetBookFunc != nil {
		return m.GetBookFunc(ctx, id)
	}
	return db.Book{}, nil
}

func (m *MockQueries) DeleteBook(ctx context.Context, id int32) error {
	if m.DeleteBookFunc != nil {
		return m.DeleteBookFunc(ctx, id)
	}
	return nil
}

func TestCreateBook(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    CreateBookRequest
		mockResponse   db.Book
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful book creation",
			requestBody: CreateBookRequest{
				Title:             "Test Book",
				Author:            "Test Author",
				Description:       "Test Description",
				YearOfPublication: 2024,
			},
			mockResponse: db.Book{
				ID:     1,
				Title:  "Test Book",
				Author: "Test Author",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "book creation without optional fields",
			requestBody: CreateBookRequest{
				Title:  "Minimal Book",
				Author: "Author Name",
			},
			mockResponse: db.Book{
				ID:     2,
				Title:  "Minimal Book",
				Author: "Author Name",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "database error",
			requestBody: CreateBookRequest{
				Title:  "Error Book",
				Author: "Error Author",
			},
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockQueries{
				CreateBookFunc: func(ctx context.Context, params db.CreateBookParams) (db.Book, error) {
					if tt.mockError != nil {
						return db.Book{}, tt.mockError
					}
					return tt.mockResponse, nil
				},
			}

			store := NewBookStore(mockDB, nil)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			store.CreateBook(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				var book db.Book
				if err := json.Unmarshal(rr.Body.Bytes(), &book); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if book.Title != tt.mockResponse.Title {
					t.Errorf("expected title %s, got %s", tt.mockResponse.Title, book.Title)
				}
			}
		})
	}
}

func TestFetchBooks(t *testing.T) {
	mockBooks := []db.Book{
		{ID: 1, Title: "Book 1", Author: "Author 1"},
		{ID: 2, Title: "Book 2", Author: "Author 2"},
	}

	mockDB := &MockQueries{
		ListBooksFunc: func(ctx context.Context) ([]db.Book, error) {
			return mockBooks, nil
		},
	}

	store := NewBookStore(mockDB, nil)
	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	rr := httptest.NewRecorder()

	store.FetchBooks(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var books []db.Book
	if err := json.Unmarshal(rr.Body.Bytes(), &books); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(books) != 2 {
		t.Errorf("expected 2 books, got %d", len(books))
	}
}

// Test FetchBooks with database error
func TestFetchBooks_DatabaseError(t *testing.T) {
	mockDB := &MockQueries{
		ListBooksFunc: func(ctx context.Context) ([]db.Book, error) {
			return nil, errors.New("database error")
		},
	}

	store := NewBookStore(mockDB, nil)
	req := httptest.NewRequest(http.MethodGet, "/books", nil)
	rr := httptest.NewRecorder()

	store.FetchBooks(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestDeleteBook(t *testing.T) {
	tests := []struct {
		name           string
		bookID         string
		getBookError   error
		deleteError    error
		expectedStatus int
	}{
		{
			name:           "successful deletion",
			bookID:         "1",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "book not found",
			bookID:         "999",
			getBookError:   pgx.ErrNoRows,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid book ID",
			bookID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "delete error",
			bookID:         "1",
			deleteError:    errors.New("delete failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockQueries{
				GetBookFunc: func(ctx context.Context, id int32) (db.Book, error) {
					if tt.getBookError != nil {
						return db.Book{}, tt.getBookError
					}
					return db.Book{ID: id, Title: "Test Book", Author: "Test Author"}, nil
				},
				DeleteBookFunc: func(ctx context.Context, id int32) error {
					return tt.deleteError
				},
			}

			store := NewBookStore(mockDB, nil)
			req := httptest.NewRequest(http.MethodDelete, "/books/"+tt.bookID, nil)
			rr := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.bookID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			store.DeleteBook(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

// Test BorrowBook handler - Note: This is a simplified test since full transaction testing
// requires a real database. For production, consider integration tests.
func TestBorrowBook_InvalidJSON(t *testing.T) {
	mockDB := &MockQueries{}
	store := NewBookStore(mockDB, nil)

	req := httptest.NewRequest(http.MethodPost, "/books/borrow", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	store.BorrowBook(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// Test ReturnBook handler
func TestReturnBook_InvalidJSON(t *testing.T) {
	mockDB := &MockQueries{}
	store := NewBookStore(mockDB, nil)

	req := httptest.NewRequest(http.MethodPost, "/books/return", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	store.ReturnBook(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// Test GetUserBorrowedBooks handler
func TestGetUserBorrowedBooks(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockBooks      []db.GetUserBorrowedBooksRow
		mockError      error
		expectedStatus int
		expectedCount  int
	}{
		{
			name:   "fetch user borrowed books",
			userID: "1",
			mockBooks: []db.GetUserBorrowedBooksRow{
				{
					ID:         1,
					BookID:     1,
					UserID:     1,
					Title:      "Book 1",
					Author:     "Author 1",
					BorrowedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
					DueDate:    pgtype.Timestamp{Time: time.Now().Add(14 * 24 * time.Hour), Valid: true},
					ReturnedAt: pgtype.Timestamp{Valid: false},
				},
				{
					ID:         2,
					BookID:     2,
					UserID:     1,
					Title:      "Book 2",
					Author:     "Author 2",
					BorrowedAt: pgtype.Timestamp{Time: time.Now().Add(-30 * 24 * time.Hour), Valid: true},
					DueDate:    pgtype.Timestamp{Time: time.Now().Add(-16 * 24 * time.Hour), Valid: true},
					ReturnedAt: pgtype.Timestamp{Valid: false},
				},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "no borrowed books",
			userID:         "1",
			mockBooks:      []db.GetUserBorrowedBooksRow{},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "database error",
			userID:         "1",
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockQueries{
				GetUserBorrowedBooksFunc: func(ctx context.Context, userID int32) ([]db.GetUserBorrowedBooksRow, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockBooks, nil
				},
			}

			store := NewBookStore(mockDB, nil)
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID+"/borrowed", nil)
			rr := httptest.NewRecorder()

			// Add chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			store.GetUserBorrowedBooks(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var books []db.GetUserBorrowedBooksRow
				if err := json.Unmarshal(rr.Body.Bytes(), &books); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if len(books) != tt.expectedCount {
					t.Errorf("expected %d books, got %d", tt.expectedCount, len(books))
				}
			}
		})
	}
}

// Test GetOverdueBooks handler
func TestGetOverdueBooks(t *testing.T) {
	tests := []struct {
		name           string
		mockBooks      []db.GetOverdueBooksRow
		mockError      error
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "fetch overdue books",
			mockBooks: []db.GetOverdueBooksRow{
				{
					ID:         1,
					BookID:     1,
					UserID:     1,
					Title:      "Overdue Book 1",
					Author:     "Author 1",
					Name:       "John Doe",
					Email:      "john@example.com",
					BorrowedAt: pgtype.Timestamp{Time: time.Now().Add(-30 * 24 * time.Hour), Valid: true},
					DueDate:    pgtype.Timestamp{Time: time.Now().Add(-2 * 24 * time.Hour), Valid: true},
					ReturnedAt: pgtype.Timestamp{Valid: false},
				},
				{
					ID:         2,
					BookID:     2,
					UserID:     2,
					Title:      "Overdue Book 2",
					Author:     "Author 2",
					Name:       "Jane Smith",
					Email:      "jane@example.com",
					BorrowedAt: pgtype.Timestamp{Time: time.Now().Add(-45 * 24 * time.Hour), Valid: true},
					DueDate:    pgtype.Timestamp{Time: time.Now().Add(-10 * 24 * time.Hour), Valid: true},
					ReturnedAt: pgtype.Timestamp{Valid: false},
				},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "no overdue books",
			mockBooks:      []db.GetOverdueBooksRow{},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "database error",
			mockError:      errors.New("database connection failed"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockQueries{
				GetOverdueBooksFunc: func(ctx context.Context) ([]db.GetOverdueBooksRow, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockBooks, nil
				},
			}

			store := NewBookStore(mockDB, nil)
			req := httptest.NewRequest(http.MethodGet, "/books/overdue", nil)
			rr := httptest.NewRecorder()

			store.GetOverdueBooks(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var books []db.GetOverdueBooksRow
				if err := json.Unmarshal(rr.Body.Bytes(), &books); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if len(books) != tt.expectedCount {
					t.Errorf("expected %d books, got %d", tt.expectedCount, len(books))
				}

				// Verify structure of first book if available
				if tt.expectedCount > 0 && len(books) > 0 {
					if books[0].Title != tt.mockBooks[0].Title {
						t.Errorf("expected title %s, got %s", tt.mockBooks[0].Title, books[0].Title)
					}
					if books[0].Name != tt.mockBooks[0].Name {
						t.Errorf("expected user name %s, got %s", tt.mockBooks[0].Name, books[0].Name)
					}
				}
			}
		})
	}
}
