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

type CreateBookRequest struct {
	Title             string `json:"title"`
	Author            string `json:"author"`
	Description       string `json:"description"`
	YearOfPublication int32  `json:"year_of_publication"`
}

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
	queries := db.New(pool)

	// Create the bookstore
	bookstore := api.NewBookStore(queries, pool)
	userstore := api.NewUserStore(queries)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	setupBookRoutes(r, bookstore)
	setupUserRoutes(r, userstore)
	setupBorrowRoutes(r, bookstore)

	srv := &http.Server{
		Addr:    ":3000",
		Handler: r,
	}

	go func() {
		log.Println("Server starting on port 3000")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exited gracefully")

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

func setupUserRoutes(r *chi.Mux, store *api.UserStore) {
	r.Get("/users", store.FetchUsers)
	r.Get("/users/{id}", store.FetchUserById)
	r.Post("/users", store.CreateUser)
	r.Delete("/users/{id}", store.DeleteUser)
}

func setupBorrowRoutes(r *chi.Mux, store *api.BookStore) {
	r.Post("/borrow", store.BorrowBook)
	r.Post("/return", store.ReturnBook)
	r.Get("/users/{id}/borrowed", store.GetUserBorrowedBooks)
	r.Get("/overdue", store.GetOverdueBooks)
}
