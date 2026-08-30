package api

import (
	"bookbackend/internal/api/models"
	db "bookbackend/internal/database"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func validateUser(name, email string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(email) == "" {
		return errors.New("email is required")
	}
	return nil
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (c CreateUserRequest) Validate() error {
	return validateUser(c.Name, c.Email)
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (u UpdateUserRequest) Validate() error {
	return validateUser(u.Name, u.Email)
}

type UserHandler struct {
	db db.Querier
}

func NewUserHandler(db db.Querier) *UserHandler {
	return &UserHandler{db: db}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with the provided details
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User details"
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [post]
func (handler *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := handler.db.CreateUser(r.Context(), db.CreateUserParams{
		Name:  req.Name,
		Email: req.Email,
	})

	if err != nil {
		var pgtrr *pgconn.PgError
		if errors.As(err, &pgtrr) && pgtrr.Code == "23505" {
			writeError(w, http.StatusConflict, "Email already exists")
			return
		}
		log.Printf("Error creating user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}
	writeJSON(w, http.StatusCreated, models.ToUserResponse(user))
}

// FetchAllUsers godoc
// @Summary List all users
// @Description Get a list of all users in the database
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} models.UserResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users [get]
func (handler *UserHandler) FetchAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := handler.db.ListUsers(r.Context())
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = models.ToUserResponse(user)
	}

	writeJSON(w, http.StatusOK, userResponses)
}

// FetchUserById godoc
// @Summary Get a user by ID
// @Description Get a single user by its ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [get]
func (handler *UserHandler) FetchUserById(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		log.Printf("Error parsing ID: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := handler.db.GetUser(r.Context(), id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Printf("Error fetching user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to fetch user")
		return
	}
	writeJSON(w, http.StatusOK, models.ToUserResponse(user))
}

// UpdateUser godoc
// @Summary Update a user by ID
// @Description Update a user's details by its ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body UpdateUserRequest true "Updated user details"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [put]
func (handler *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		log.Printf("Error parsing ID: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update the user
	user, err := handler.db.UpdateUser(r.Context(), db.UpdateUserParams{
		ID:    id,
		Name:  req.Name,
		Email: req.Email,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Printf("Error updating user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	writeJSON(w, http.StatusOK, models.ToUserResponse(user))
}

// DeleteUser godoc
// @Summary Delete a user by ID
// @Description Delete a user by its ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 204 "User deleted successfully"
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id} [delete]
func (handler *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		log.Printf("Error parsing ID: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	_, err = handler.db.DeleteUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Printf("Error deleting user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// BorrowHistory godoc
// @Summary Get a user's borrow history
// @Description Get the borrow history of a user by its ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {array} models.HistoryRowResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /users/{id}/history [get]
func (handler *UserHandler) BorrowHistory(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		log.Printf("Error parsing ID: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	books, err := handler.db.GetUserBorrowHistory(r.Context(), id)
	if err != nil {
		log.Printf("Error fetching borrow history: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to fetch borrow history")
		return
	}
	history := make([]models.HistoryRowResponse, len(books))
	for i, book := range books {
		history[i] = models.ToHistoryRowResponse(book)
	}
	writeJSON(w, http.StatusOK, history)
}

func (handler *UserHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", handler.FetchAllUsers)
	r.Get("/{id}", handler.FetchUserById)
	r.Get("/{id}/history", handler.BorrowHistory)
	r.Post("/", handler.CreateUser)
	r.Put("/{id}", handler.UpdateUser)
	r.Delete("/{id}", handler.DeleteUser)
	return r
}
