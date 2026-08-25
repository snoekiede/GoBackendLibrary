package models

import (
	db "bookbackend/internal/database"
	"time"
)

type UserResponse struct {
	ID        int32     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToUserResponse(user db.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}
}
