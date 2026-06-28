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

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

// MockUserQueries extends MockQueries with user-specific methods
type MockUserQueries struct {
	db.Querier
	CreateUserFunc func(ctx context.Context, params db.CreateUserParams) (db.User, error)
	ListUsersFunc  func(ctx context.Context) ([]db.User, error)
	GetUserFunc    func(ctx context.Context, id int32) (db.User, error)
	UpdateUserFunc func(ctx context.Context, params db.UpdateUserParams) (db.User, error)
	DeleteUserFunc func(ctx context.Context, id int32) error
}

func (m *MockUserQueries) WithTx(tx pgx.Tx) *db.Queries {
	return db.New(tx)
}

func (m *MockUserQueries) CreateUser(ctx context.Context, params db.CreateUserParams) (db.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, params)
	}
	return db.User{}, nil
}

func (m *MockUserQueries) ListUsers(ctx context.Context) ([]db.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx)
	}
	return []db.User{}, nil
}

func (m *MockUserQueries) GetUser(ctx context.Context, id int32) (db.User, error) {
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, id)
	}
	return db.User{}, nil
}

func (m *MockUserQueries) UpdateUser(ctx context.Context, params db.UpdateUserParams) (db.User, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, params)
	}
	return db.User{}, nil
}

func (m *MockUserQueries) DeleteUser(ctx context.Context, id int32) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}

// Test CreateUser handler
func TestCreateUser(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    CreateUserRequest
		mockResponse   db.User
		mockError      error
		expectedStatus int
	}{
		{
			name: "successful user creation",
			requestBody: CreateUserRequest{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			mockResponse: db.User{
				ID:    1,
				Name:  "John Doe",
				Email: "john@example.com",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "user creation with minimal data",
			requestBody: CreateUserRequest{
				Name:  "Jane Smith",
				Email: "jane@example.com",
			},
			mockResponse: db.User{
				ID:    2,
				Name:  "Jane Smith",
				Email: "jane@example.com",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "database error",
			requestBody: CreateUserRequest{
				Name:  "Error User",
				Email: "error@example.com",
			},
			mockError:      errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockUserQueries{
				CreateUserFunc: func(ctx context.Context, params db.CreateUserParams) (db.User, error) {
					if tt.mockError != nil {
						return db.User{}, tt.mockError
					}
					return tt.mockResponse, nil
				},
			}

			store := NewUserStore(mockDB)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			store.CreateUser(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				var user db.User
				json.Unmarshal(rr.Body.Bytes(), &user)
				if user.Email != tt.mockResponse.Email {
					t.Errorf("expected email %s, got %s", tt.mockResponse.Email, user.Email)
				}
				if user.Name != tt.mockResponse.Name {
					t.Errorf("expected name %s, got %s", tt.mockResponse.Name, user.Name)
				}
			}
		})
	}
}

// Test invalid JSON in CreateUser
func TestCreateUser_InvalidJSON(t *testing.T) {
	mockDB := &MockUserQueries{}
	store := NewUserStore(mockDB)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	store.CreateUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// Test FetchUsers handler
func TestFetchUsers(t *testing.T) {
	tests := []struct {
		name           string
		mockUsers      []db.User
		mockError      error
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "fetch multiple users",
			mockUsers: []db.User{
				{ID: 1, Name: "User 1", Email: "user1@example.com"},
				{ID: 2, Name: "User 2", Email: "user2@example.com"},
				{ID: 3, Name: "User 3", Email: "user3@example.com"},
			},
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
		{
			name:           "fetch empty list",
			mockUsers:      []db.User{},
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
			mockDB := &MockUserQueries{
				ListUsersFunc: func(ctx context.Context) ([]db.User, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return tt.mockUsers, nil
				},
			}

			store := NewUserStore(mockDB)
			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			rr := httptest.NewRecorder()

			store.FetchUsers(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var users []db.User
				json.Unmarshal(rr.Body.Bytes(), &users)

				if len(users) != tt.expectedCount {
					t.Errorf("expected %d users, got %d", tt.expectedCount, len(users))
				}
			}
		})
	}
}

// Test FetchUserById handler
func TestFetchUserById(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockResponse   db.User
		mockError      error
		expectedStatus int
	}{
		{
			name:   "fetch existing user",
			userID: "1",
			mockResponse: db.User{
				ID:    1,
				Name:  "John Doe",
				Email: "john@example.com",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user not found",
			userID:         "999",
			mockError:      pgx.ErrNoRows,
			expectedStatus: http.StatusNotFound,
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
			mockDB := &MockUserQueries{
				GetUserFunc: func(ctx context.Context, id int32) (db.User, error) {
					if tt.mockError != nil {
						return db.User{}, tt.mockError
					}
					return tt.mockResponse, nil
				},
			}

			store := NewUserStore(mockDB)
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)
			rr := httptest.NewRecorder()

			// Add chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			store.FetchUserById(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var user db.User
				json.Unmarshal(rr.Body.Bytes(), &user)
				if user.ID != tt.mockResponse.ID {
					t.Errorf("expected user ID %d, got %d", tt.mockResponse.ID, user.ID)
				}
			}
		})
	}
}

// Test UpdateUser handler
func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    UpdateUserRequest
		getUserError   error
		updateResponse db.User
		updateError    error
		expectedStatus int
	}{
		{
			name:   "successful user update",
			userID: "1",
			requestBody: UpdateUserRequest{
				Name:  "Updated Name",
				Email: "updated@example.com",
			},
			updateResponse: db.User{
				ID:    1,
				Name:  "Updated Name",
				Email: "updated@example.com",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "user not found for update",
			userID: "999",
			requestBody: UpdateUserRequest{
				Name:  "Test Name",
				Email: "test@example.com",
			},
			getUserError:   pgx.ErrNoRows,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "update error",
			userID: "1",
			requestBody: UpdateUserRequest{
				Name:  "Test Name",
				Email: "test@example.com",
			},
			updateError:    errors.New("update failed"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "get user database error",
			userID:         "1",
			getUserError:   errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockUserQueries{
				GetUserFunc: func(ctx context.Context, id int32) (db.User, error) {
					if tt.getUserError != nil {
						return db.User{}, tt.getUserError
					}
					return db.User{ID: id, Name: "Old Name", Email: "old@example.com"}, nil
				},
				UpdateUserFunc: func(ctx context.Context, params db.UpdateUserParams) (db.User, error) {
					if tt.updateError != nil {
						return db.User{}, tt.updateError
					}
					return tt.updateResponse, nil
				},
			}

			store := NewUserStore(mockDB)

			var req *http.Request
			if tt.expectedStatus == http.StatusBadRequest && tt.userID == "invalid" {
				req = httptest.NewRequest(http.MethodPut, "/users/"+tt.userID, nil)
			} else {
				body, _ := json.Marshal(tt.requestBody)
				req = httptest.NewRequest(http.MethodPut, "/users/"+tt.userID, bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
			}

			rr := httptest.NewRecorder()

			// Add chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			store.UpdateUser(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var user db.User
				json.Unmarshal(rr.Body.Bytes(), &user)
				if user.Name != tt.updateResponse.Name {
					t.Errorf("expected name %s, got %s", tt.updateResponse.Name, user.Name)
				}
			}
		})
	}
}

// Test UpdateUser with invalid JSON
func TestUpdateUser_InvalidJSON(t *testing.T) {
	mockDB := &MockUserQueries{}
	store := NewUserStore(mockDB)

	req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	store.UpdateUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// Test DeleteUser handler
func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		getUserError   error
		deleteError    error
		expectedStatus int
	}{
		{
			name:           "successful user deletion",
			userID:         "1",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "user not found for deletion",
			userID:         "999",
			getUserError:   pgx.ErrNoRows,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "delete error",
			userID:         "1",
			deleteError:    errors.New("delete failed"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "get user database error",
			userID:         "1",
			getUserError:   errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := &MockUserQueries{
				GetUserFunc: func(ctx context.Context, id int32) (db.User, error) {
					if tt.getUserError != nil {
						return db.User{}, tt.getUserError
					}
					return db.User{ID: id, Name: "Test User", Email: "test@example.com"}, nil
				},
				DeleteUserFunc: func(ctx context.Context, id int32) error {
					return tt.deleteError
				},
			}

			store := NewUserStore(mockDB)
			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.userID, nil)
			rr := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			store.DeleteUser(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
