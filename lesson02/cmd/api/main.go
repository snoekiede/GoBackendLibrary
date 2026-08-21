package main

import (
	db "bookbackend/internal/database"
	"context"
	"log"
	"os"

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
		log.Fatalf("Unable to create a connection pool: %v", err)
	}

	defer pool.Close()

	// test if we can reach the database
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to connect to the database: %v", err)
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

}
