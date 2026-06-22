package main

import (
	"bookbackend/internal/api"
	db "bookbackend/internal/database"
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type CreateBookRequest struct {
	Title             string `json:"title"`
	Author            string `json:"author"`
	Description       string `json:"description"`
	YearOfPublication int32  `json:"year_of_publication"`
}

func main() {
	//get the connection from an environment variable

	godotenv.Load()

	dbUrl := os.Getenv("DATABASE_URL")

	if dbUrl == "" {
		panic("DATABASE_URL environment variable is not set")
	}

	pool, err := pgxpool.New(context.Background(), dbUrl)
	queries := db.New(pool)

	// Create the bookstore
	store := api.NewBookStore(queries)

	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	setupBookRoutes(r, store)

	http.ListenAndServe(":3000", r)
}

func setupBookRoutes(r *chi.Mux, store *api.BookStore) {
	r.Get("/books", func(w http.ResponseWriter, r *http.Request) {
		store.FetchBooks(w, r)
	})

	r.Get("/books/{id}", func(w http.ResponseWriter, r *http.Request) {
		store.FetchBookByID(w, r)
	})

	r.Post("/books", func(w http.ResponseWriter, r *http.Request) {
		store.CreateBook(w, r)
	})

	r.Delete("/books/{id}", func(w http.ResponseWriter, r *http.Request) {
		store.DeleteBook(w, r)
	})
}
