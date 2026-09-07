package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"book-catalog-api/graph"
	"book-catalog-api/internal/database"
	"book-catalog-api/internal/repository/postgres"
	"book-catalog-api/internal/rest"

	"github.com/joho/godotenv"

	"github.com/99designs/gqlgen/graphql/handler"
	// "github.com/99designs/gqlgen/graphql/handler/extension"
	// "github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
)

func main() {
	ctx := context.Background()

	// Загружаем переменные из .env
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using environment variables")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Подключение к PostgreSQL
	db, err := database.NewPostgresPool(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Connected to PostgreSQL")

	// Репозитории
	bookRepository := postgres.NewBookRepository(db)
	authorRepository := postgres.NewAuthorRepository(db)
	userRepository := postgres.NewUserRepository(db)

	graphqlResolver := &graph.Resolver{
		BookRepository:   bookRepository,
		AuthorRepository: authorRepository,
	}

	graphqlServer := handler.NewDefaultServer(
		graph.NewExecutableSchema(
			graph.Config{
				Resolvers: graphqlResolver,
			},
		),
	)

	// REST handlers
	bookHandler := rest.NewBookHandler(bookRepository)
	authorHandler := rest.NewAuthorHandler(authorRepository)
	userHandler := rest.NewUserHandler(userRepository)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Books
	mux.HandleFunc("GET /books", bookHandler.GetAll)
	mux.HandleFunc("GET /books/{id}", bookHandler.GetByID)
	mux.HandleFunc("POST /books", bookHandler.Create)
	mux.HandleFunc("PUT /books/{id}", bookHandler.Update)
	mux.HandleFunc("DELETE /books/{id}", bookHandler.Delete)

	// Authors
	mux.HandleFunc("GET /authors", authorHandler.GetAll)
	mux.HandleFunc("GET /authors/{id}", authorHandler.GetByID)
	mux.HandleFunc("POST /authors", authorHandler.Create)
	mux.HandleFunc("PUT /authors/{id}", authorHandler.Update)
	mux.HandleFunc("DELETE /authors/{id}", authorHandler.Delete)

	// Users
	mux.HandleFunc("GET /users", userHandler.GetAll)
	mux.HandleFunc("GET /users/{id}", userHandler.GetByID)

	// Reading list
	mux.HandleFunc(
		"GET /users/{id}/reading-list",
		userHandler.GetReadingList,
	)

	mux.HandleFunc(
		"POST /users/{id}/reading-list",
		userHandler.AddToReadingList,
	)

	mux.HandleFunc(
		"DELETE /users/{id}/reading-list/{book_id}",
		userHandler.RemoveFromReadingList,
	)

	mux.Handle(
		"/graphql",
		graphqlServer,
	)

	mux.Handle(
		"/playground",
		playground.Handler(
			"GraphQL Playground",
			"/graphql",
		),
	)

	// Frontend
	fileServer := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fileServer)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fmt.Printf("Server started on http://localhost:%s\n", port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
