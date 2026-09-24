package main

import (
	"bookbackend/internal/api"
	db "bookbackend/internal/database"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "bookbackend/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title Library API
// @version 1.0
// @description API for managing a library with borrowing functionality
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@bookstore.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return errors.New("DATABASE_URL environment variable is not set")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
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

	serverErr := make(chan error, 1)

	go func() {
		log.Println("Server is running on port 3000")

		if err := srv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	var serveErr error

	select {
	case sig := <-quit:
		log.Printf("Received %s, shutting down server...", sig)

	case serveErr = <-serverErr:
		log.Printf("Server stopped unexpectedly: %v", serveErr)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)

		if closeErr := srv.Close(); closeErr != nil {
			return fmt.Errorf(
				"graceful shutdown failed: %v; forced close failed: %w",
				err,
				closeErr,
			)
		}
	}

	log.Println("Server stopped")

	if serveErr != nil {
		return fmt.Errorf("HTTP server failed: %w", serveErr)
	}

	return nil

}
