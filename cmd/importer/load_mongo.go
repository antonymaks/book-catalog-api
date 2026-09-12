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

type mongoUser struct {
	ID           int64  `bson:"id"`
	Username     string `bson:"username"`
	PasswordHash string `bson:"password_hash"`
	Role         string `bson:"role"`
}

type mongoReadingList struct {
	UserID int64 `bson:"user_id"`
	BookID int64 `bson:"book_id"`
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
	usersCollection := db.Collection("users")
	readingListCollection := db.Collection("reading_list")

	file, err := os.Open(mongoInputFile)
	if err != nil {
		panic(fmt.Errorf("failed to open csv: %w", err))
	}

	defer file.Close()

	reader := csv.NewReader(file)

	if _, err := reader.Read(); err != nil {
		panic(fmt.Errorf("failed to read header: %w", err))
	}

	// Полностью очищаем данные перед повторным импортом.
	collections := []*mongo.Collection{
		readingListCollection,
		usersCollection,
		booksCollection,
		authorsCollection,
	}

	for _, collection := range collections {
		if _, err := collection.DeleteMany(ctx, bson.M{}); err != nil {
			panic(fmt.Errorf(
				"failed to clear collection %s: %w",
				collection.Name(),
				err,
			))
		}
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

	users := []interface{}{
		mongoUser{
			ID:           1,
			Username:     "admin",
			PasswordHash: "test_hash_admin",
			Role:         "admin",
		},
		mongoUser{
			ID:           2,
			Username:     "anton",
			PasswordHash: "test_hash_anton",
			Role:         "user",
		},
		mongoUser{
			ID:           3,
			Username:     "alex",
			PasswordHash: "test_hash_alex",
			Role:         "user",
		},
	}

	if _, err := usersCollection.InsertMany(ctx, users); err != nil {
		panic(fmt.Errorf("failed to insert users: %w", err))
	}

	// Небольшой тестовый список чтения.
	// Используем существующие ID книг из нового датасета.
	readingList := []interface{}{
		mongoReadingList{
			UserID: 2,
			BookID: 1,
		},
		mongoReadingList{
			UserID: 2,
			BookID: 2,
		},
		mongoReadingList{
			UserID: 3,
			BookID: 3,
		},
	}

	if _, err := readingListCollection.InsertMany(ctx, readingList); err != nil {
		panic(fmt.Errorf(
			"failed to insert reading list: %w",
			err,
		))
	}

	createIndexes(ctx, db)

	fmt.Println("MongoDB import completed")
	fmt.Printf("Authors: %d\n", len(authors))
	fmt.Printf("Books: %d\n", len(books))
	fmt.Printf("Users: %d\n", len(users))
	fmt.Printf("Reading list items: %d\n", len(readingList))
}

func createIndexes(
	ctx context.Context,
	db *mongo.Database,
) {
	authors := db.Collection("authors")
	books := db.Collection("books")
	users := db.Collection("users")
	readingList := db.Collection("reading_list")

	indexes := []struct {
		collection *mongo.Collection
		model      mongo.IndexModel
	}{
		{
			collection: authors,
			model: mongo.IndexModel{
				Keys: bson.D{{"id", 1}},
				Options: options.Index().
					SetUnique(true),
			},
		},
		{
			collection: books,
			model: mongo.IndexModel{
				Keys: bson.D{{"id", 1}},
				Options: options.Index().
					SetUnique(true),
			},
		},
		{
			collection: books,
			model: mongo.IndexModel{
				Keys: bson.D{{"author_id", 1}},
			},
		},
		{
			collection: users,
			model: mongo.IndexModel{
				Keys: bson.D{{"id", 1}},
				Options: options.Index().
					SetUnique(true),
			},
		},
		{
			collection: users,
			model: mongo.IndexModel{
				Keys: bson.D{{"username", 1}},
				Options: options.Index().
					SetUnique(true),
			},
		},
		{
			collection: readingList,
			model: mongo.IndexModel{
				Keys: bson.D{
					{"user_id", 1},
					{"book_id", 1},
				},
				Options: options.Index().
					SetUnique(true),
			},
		},
	}

	for _, index := range indexes {
		if _, err := index.collection.Indexes().
			CreateOne(ctx, index.model); err != nil {

			panic(fmt.Errorf(
				"failed to create index for %s: %w",
				index.collection.Name(),
				err,
			))
		}
	}
}
