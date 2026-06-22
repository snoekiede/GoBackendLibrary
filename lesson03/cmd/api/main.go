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

func main() {
	//get the connection from an environment variable

	godotenv.Load()

	dbUrl := os.Getenv("DATABASE_URL")

	if dbUrl == "" {
		panic("DATABASE_URL environment variable is not set")
	}

	pool, err := pgxpool.New(context.Background(), dbUrl)

	queries := db.New(pool)
	bookhandler := api.NewBookHandler(queries)

	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	addBookRoutes(r, bookhandler)

	http.ListenAndServe(":3000", r)
}

func addBookRoutes(router *chi.Mux, bookhandler *api.BookHandler) {
	router.Get("/books", func(w http.ResponseWriter, r *http.Request) {
		bookhandler.FetchBooks(w, r)
	})

	router.Get("/books/{id}", func(w http.ResponseWriter, r *http.Request) {
		bookhandler.FetchBookByID(w, r)
	})

	router.Post("/books", func(w http.ResponseWriter, r *http.Request) {
		bookhandler.CreateBook(w, r)
	})

	router.Delete("/books/{id}", func(w http.ResponseWriter, r *http.Request) {
		bookhandler.DeleteBook(w, r)
	})
}
