package main

import (
	_ "bookbackend/docs"
	"bookbackend/internal/api"
	db "bookbackend/internal/database"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	srv := &http.Server{
		Addr:              ":3000",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Println("Server is running on port 3000")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Graceful shutdown failed %v:", err)
		if err := srv.Close(); err != nil {
			log.Fatalf("Unable to close server: %v", err)
		}
	}

	log.Println("Server stopped")
}
