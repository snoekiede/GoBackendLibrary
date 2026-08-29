package main

import (
	_ "bookbackend/docs"
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
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Library API
// @version 1.0
// @description API for managing a book store with borrowing functionality
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@bookstore.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /
// @schemes http
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
	bookhandler := api.NewBookHandler(queries, pool)
	userhandler := api.NewUserHandler(queries)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Mount("/books", bookhandler.Routes())
	r.Mount("/users", userhandler.Routes())
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	log.Fatal(http.ListenAndServe(":3000", r))
}
