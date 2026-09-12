package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"book-catalog-api/graph"
	"book-catalog-api/internal/auth"
	"book-catalog-api/internal/database"
	"book-catalog-api/internal/repository"
	"book-catalog-api/internal/rest"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"

	mongorepo "book-catalog-api/internal/repository/mongo"
	postgresrepo "book-catalog-api/internal/repository/postgres"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using environment variables")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		dbDriver = "postgres"
	}

	var bookRepository repository.BookRepository
	var authorRepository repository.AuthorRepository
	var userRepository repository.UserRepository

	switch dbDriver {

	case "postgres":
		databaseURL := os.Getenv("DATABASE_URL")

		if databaseURL == "" {
			log.Fatal("DATABASE_URL is not set")
		}

		db, err := database.NewPostgresPool(
			ctx,
			databaseURL,
		)
		if err != nil {
			log.Fatal(err)
		}

		defer db.Close()

		bookRepository = postgresrepo.NewBookRepository(db)
		authorRepository = postgresrepo.NewAuthorRepository(db)
		userRepository = postgresrepo.NewUserRepository(db)

		fmt.Println("Connected to PostgreSQL")

	case "mongo":
		mongoURL := os.Getenv("MONGO_URL")
		databaseName := os.Getenv("MONGO_DATABASE")

		if mongoURL == "" {
			log.Fatal("MONGO_URL is not set")
		}

		if databaseName == "" {
			log.Fatal("MONGO_DATABASE is not set")
		}

		client, err := database.NewMongoClient(
			ctx,
			mongoURL,
		)
		if err != nil {
			log.Fatal(err)
		}

		defer client.Disconnect(ctx)

		db := client.Database(databaseName)

		bookRepository = mongorepo.NewBookRepository(db)
		authorRepository = mongorepo.NewAuthorRepository(db)
		userRepository = mongorepo.NewUserRepository(db)

		fmt.Println("Connected to MongoDB")

	default:
		log.Fatalf(
			"unsupported DB_DRIVER: %s",
			dbDriver,
		)
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	authService := auth.NewService(
		userRepository,
		jwtSecret,
	)

	authHandler := rest.NewAuthHandler(
		authService,
		userRepository,
	)

	graphqlResolver := &graph.Resolver{
		BookRepository:   bookRepository,
		AuthorRepository: authorRepository,
		UserRepository:   userRepository,
	}

	graphqlServer := handler.NewDefaultServer(
		graph.NewExecutableSchema(
			graph.Config{
				Resolvers: graphqlResolver,
			},
		),
	)

	bookHandler := rest.NewBookHandler(bookRepository)
	authorHandler := rest.NewAuthorHandler(authorRepository)
	userHandler := rest.NewUserHandler(userRepository)

	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /health",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusOK)

			fmt.Fprintf(
				w,
				`{"status":"ok","database":"%s"}`,
				dbDriver,
			)
		},
	)

	// Books
	mux.HandleFunc(
		"GET /books",
		bookHandler.GetAll,
	)

	mux.HandleFunc(
		"GET /books/{id}",
		bookHandler.GetByID,
	)

	mux.Handle(
		"POST /books",
		authService.RequireRole(
			"admin",
			http.HandlerFunc(
				bookHandler.Create,
			),
		),
	)

	mux.Handle(
		"PUT /books/{id}",
		authService.RequireRole(
			"admin",
			http.HandlerFunc(
				bookHandler.Update,
			),
		),
	)

	mux.Handle(
		"DELETE /books/{id}",
		authService.RequireRole(
			"admin",
			http.HandlerFunc(
				bookHandler.Delete,
			),
		),
	)

	// Authors
	mux.HandleFunc(
		"GET /authors",
		authorHandler.GetAll,
	)

	mux.HandleFunc(
		"GET /authors/{id}",
		authorHandler.GetByID,
	)

	mux.Handle(
		"POST /authors",
		authService.RequireRole(
			"admin",
			http.HandlerFunc(
				authorHandler.Create,
			),
		),
	)

	mux.Handle(
		"PUT /authors/{id}",
		authService.RequireRole(
			"admin",
			http.HandlerFunc(
				authorHandler.Update,
			),
		),
	)

	mux.Handle(
		"DELETE /authors/{id}",
		authService.RequireRole(
			"admin",
			http.HandlerFunc(
				authorHandler.Delete,
			),
		),
	)

	// Users
	mux.HandleFunc(
		"GET /users",
		userHandler.GetAll,
	)

	mux.HandleFunc(
		"GET /users/{id}",
		userHandler.GetByID,
	)

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

	// GraphQL
	mux.Handle(
		"/graphql",
		authService.OptionalMiddleware(
			graphqlServer,
		),
	)

	mux.Handle(
		"/playground",
		playground.Handler(
			"GraphQL Playground",
			"/graphql",
		),
	)

	mux.HandleFunc(
		"POST /auth/register",
		authHandler.Register,
	)

	mux.HandleFunc(
		"POST /auth/login",
		authHandler.Login,
	)

	mux.Handle(
		"GET /auth/me",
		authService.Middleware(
			http.HandlerFunc(
				authHandler.Me,
			),
		),
	)

	mux.Handle(
		"GET /me/reading-list",
		authService.Middleware(
			http.HandlerFunc(
				userHandler.GetMyReadingList,
			),
		),
	)

	mux.Handle(
		"POST /me/reading-list",
		authService.Middleware(
			http.HandlerFunc(
				userHandler.AddToMyReadingList,
			),
		),
	)

	mux.Handle(
		"DELETE /me/reading-list/{book_id}",
		authService.Middleware(
			http.HandlerFunc(
				userHandler.RemoveFromMyReadingList,
			),
		),
	)

	// Frontend
	fileServer := http.FileServer(
		http.Dir("./web"),
	)

	mux.Handle("/", fileServer)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMiddleware(mux),
	}

	fmt.Printf(
		"Server started on http://localhost:%s\n",
		port,
	)

	fmt.Printf(
		"Database driver: %s\n",
		dbDriver,
	)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type",
		)
		w.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, DELETE, OPTIONS",
		)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
