package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const mongoInputFile = "data/prepared/books.csv"

type mongoAuthor struct {
	ID   int64  `bson:"id"`
	Name string `bson:"name"`
}

type mongoBook struct {
	ID          int64  `bson:"id"`
	AuthorID    int64  `bson:"author_id"`
	Title       string `bson:"title"`
	Description string `bson:"description"`
}

func main() {
	_ = godotenv.Load()

	mongoURL := os.Getenv("MONGO_URL")
	databaseName := os.Getenv("MONGO_DATABASE")

	if mongoURL == "" {
		panic("MONGO_URL is not set")
	}

	if databaseName == "" {
		panic("MONGO_DATABASE is not set")
	}

	ctx := context.Background()

	client, err := mongo.Connect(
		options.Client().ApplyURI(mongoURL),
	)
	if err != nil {
		panic(fmt.Errorf("failed to connect to MongoDB: %w", err))
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		panic(fmt.Errorf("failed to ping MongoDB: %w", err))
	}

	db := client.Database(databaseName)

	authorsCollection := db.Collection("authors")
	booksCollection := db.Collection("books")

	file, err := os.Open(mongoInputFile)
	if err != nil {
		panic(fmt.Errorf("failed to open csv: %w", err))
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Пропускаем header
	if _, err := reader.Read(); err != nil {
		panic(fmt.Errorf("failed to read header: %w", err))
	}

	// Делаем импорт повторяемым:
	// каждый запуск заново формирует books и authors.
	if _, err := booksCollection.DeleteMany(ctx, bson.M{}); err != nil {
		panic(fmt.Errorf("failed to clear books: %w", err))
	}

	if _, err := authorsCollection.DeleteMany(ctx, bson.M{}); err != nil {
		panic(fmt.Errorf("failed to clear authors: %w", err))
	}

	authorIDs := make(map[string]int64)

	var authors []interface{}
	var books []interface{}

	var nextAuthorID int64 = 1
	var nextBookID int64 = 1

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			panic(fmt.Errorf("failed to read csv row: %w", err))
		}

		if len(record) < 3 {
			continue
		}

		title := strings.TrimSpace(record[0])
		authorName := strings.TrimSpace(record[1])
		description := strings.TrimSpace(record[2])

		if title == "" || authorName == "" {
			continue
		}

		authorKey := strings.ToLower(authorName)

		authorID, exists := authorIDs[authorKey]

		if !exists {
			authorID = nextAuthorID
			nextAuthorID++

			authorIDs[authorKey] = authorID

			authors = append(authors, mongoAuthor{
				ID:   authorID,
				Name: authorName,
			})
		}

		books = append(books, mongoBook{
			ID:          nextBookID,
			AuthorID:    authorID,
			Title:       title,
			Description: description,
		})

		nextBookID++
	}

	if len(authors) > 0 {
		if _, err := authorsCollection.InsertMany(ctx, authors); err != nil {
			panic(fmt.Errorf("failed to insert authors: %w", err))
		}
	}

	if len(books) > 0 {
		if _, err := booksCollection.InsertMany(ctx, books); err != nil {
			panic(fmt.Errorf("failed to insert books: %w", err))
		}
	}

	fmt.Printf("MongoDB import completed\n")
	fmt.Printf("Authors: %d\n", len(authors))
	fmt.Printf("Books: %d\n", len(books))
}