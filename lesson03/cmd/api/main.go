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

	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	queries := db.New(pool)
	bookhandler := api.NewBookHandler(queries)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	addBookRoutes(r, bookhandler)

	http.ListenAndServe(":3000", r)
}

func addBookRoutes(router *chi.Mux, bookhandler *api.BookHandler) {
	router.Get("/books", bookhandler.FetchBooks)
	router.Get("/books/{id}", bookhandler.FetchBookByID)
	router.Post("/books", bookhandler.CreateBook)
	router.Put("/books/{id}", bookhandler.UpdateBook)
	router.Delete("/books/{id}", bookhandler.DeleteBook)
}
