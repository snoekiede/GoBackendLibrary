package models

import (
	db "bookbackend/internal/database"
	"time"
)

type OverdueBooksRowResponse struct {
	ID         int32     `json:"id"`
	BookID     int32     `json:"book_id"`
	UserID     int32     `json:"user_id"`
	BorrowedAt time.Time `json:"borrowed_at"`
	DueDate    time.Time `json:"due_date"`
	ReturnedAt time.Time `json:"returned_at"`
	Title      string    `json:"title"`
	Author     string    `json:"author"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
}

func (r OverdueBooksRowResponse) ToOverdueBooksRowResponse(row db.GetOverdueBooksRow) OverdueBooksRowResponse {
	return OverdueBooksRowResponse{
		ID:         row.ID,
		BookID:     row.BookID,
		UserID:     row.UserID,
		BorrowedAt: row.BorrowedAt.Time,
		DueDate:    row.DueDate.Time,
		ReturnedAt: row.ReturnedAt.Time,
		Title:      row.Title,
		Author:     row.Author,
		Name:       row.Name,
		Email:      row.Email,
	}
}
