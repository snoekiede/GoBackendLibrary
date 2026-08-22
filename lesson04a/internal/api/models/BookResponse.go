package models

import db "bookbackend/internal/database"

type BookResponse struct {
	ID                int32  `json:"id"`
	Title             string `json:"title"`
	Author            string `json:"author"`
	Description       string `json:"description,omitempty"`
	YearOfPublication int32  `json:"year_of_publication,omitempty"`
}

func ToBookResponse(book db.Book) BookResponse {
	response := BookResponse{
		ID:     book.ID,
		Title:  book.Title,
		Author: book.Author,
	}

	if book.Description.Valid {
		response.Description = book.Description.String
	}

	if book.YearOfPublication.Valid {
		response.YearOfPublication = book.YearOfPublication.Int32
	}

	return response
}
