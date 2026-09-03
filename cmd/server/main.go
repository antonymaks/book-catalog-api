package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"book-catalog-api/internal/database"
	postgresrepo "book-catalog-api/internal/repository/postgres"
	"book-catalog-api/internal/rest"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, using environment variables")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := database.NewPostgresPool(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Connected to PostgreSQL")

	bookRepository := postgresrepo.NewBookRepository(db)
	authorRepository := postgresrepo.NewAuthorRepository(db)
	userRepository := postgresrepo.NewUserRepository(db)

	bookHandler := rest.NewBookHandler(bookRepository)
	authorHandler := rest.NewAuthorHandler(authorRepository)
	userHandler := rest.NewUserHandler(userRepository)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"Book Catalog API"}`)
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(r.Context()); err != nil {
			http.Error(
				w,
				`{"status":"database error"}`,
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	mux.HandleFunc("GET /books", bookHandler.GetAll)
	mux.HandleFunc("GET /books/{id}", bookHandler.GetByID)
	mux.HandleFunc("GET /authors", authorHandler.GetAll)
	mux.HandleFunc("GET /authors/{id}", authorHandler.GetByID)
	mux.HandleFunc("GET /users", userHandler.GetAll)
	mux.HandleFunc("GET /users/{id}", userHandler.GetByID)
	mux.HandleFunc("GET /users/{id}/reading-list", userHandler.GetReadingList)

	mux.HandleFunc("POST /books", bookHandler.Create)
	mux.HandleFunc("POST /authors", authorHandler.Create)
	mux.HandleFunc("POST /users/{id}/reading-list", userHandler.AddToReadingList)

	mux.HandleFunc("PUT /books/{id}", bookHandler.Update)
	mux.HandleFunc("PUT /authors/{id}", authorHandler.Update)

	mux.HandleFunc("DELETE /books/{id}", bookHandler.Delete)
	mux.HandleFunc("DELETE /authors/{id}", authorHandler.Delete)
	mux.HandleFunc("DELETE /users/{id}/reading-list/{book_id}", userHandler.RemoveFromReadingList)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fmt.Printf("Server started on http://localhost:%s\n", port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
