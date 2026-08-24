package api

import (
	"bookbackend/internal/api/models"
	db "bookbackend/internal/database"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

func validateUser(name, email string) error {
	if name == "" {
		return errors.New("name is required")
	}
	if email == "" {
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
		log.Printf("Error creating user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}
	writeJSON(w, http.StatusCreated, models.ToUserResponse(user))
}

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

	// First check if user exists
	_, err = handler.db.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Printf("Error fetching user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to fetch user")
		return
	}

	// Update the user
	user, err := handler.db.UpdateUser(r.Context(), db.UpdateUserParams{
		ID:    id,
		Name:  req.Name,
		Email: req.Email,
	})

	if err != nil {
		log.Printf("Error updating user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}

	writeJSON(w, http.StatusOK, models.ToUserResponse(user))
}

func (handler *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		log.Printf("Error parsing ID: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Check if user exists before attempting to delete
	_, err = handler.db.GetUser(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Printf("Error fetching user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to fetch user")
		return
	}

	err = handler.db.DeleteUser(r.Context(), int32(id))
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (handler *UserHandler) SetupRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", handler.FetchAllUsers)
	r.Get("/{id}", handler.FetchUserById)
	r.Post("/", handler.CreateUser)
	r.Put("/{id}", handler.UpdateUser)
	r.Delete("/{id}", handler.DeleteUser)
	return r
}
