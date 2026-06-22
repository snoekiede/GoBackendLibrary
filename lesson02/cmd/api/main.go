package main

import (
	db "bookbackend/internal/database"
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	//get the connection from an environment variable

	godotenv.Load()

	dbUrl := os.Getenv("DATABASE_URL")

	if dbUrl == "" {

		panic("DATABASE_URL environment variable is not set")
	}

	pool, err := pgxpool.New(context.Background(), dbUrl)

	if err != nil {
		panic(err)
	}
	
	queries := db.New(pool)

	book, err := queries.CreateBook(context.Background(), db.CreateBookParams{
		Title:  "The Go Programming Language",
		Author: "Alan Donovan",
		Description: pgtype.Text{
			String: "A comprehensive guide to Go",
			Valid:  true,
		},
		YearOfPublication: pgtype.Int4{
			Int32: 2015,
			Valid: true,
		},
	})
	if err != nil {
		log.Printf("Error creating book: %v", err)
	} else {
		log.Printf("Created book: ID=%d, Title=%s, Author=%s", book.ID, book.Title, book.Author)
	}

	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	defer pool.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	http.ListenAndServe(":3000", r)
}
