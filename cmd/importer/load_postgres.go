package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"book-catalog-api/internal/database"
	"book-catalog-api/internal/domain"
	"book-catalog-api/internal/repository/postgres"

	"github.com/joho/godotenv"
)

const inputFile = "data/prepared/books.csv"

func main() {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		panic("DATABASE_URL is not set")
	}

	ctx := context.Background()

	db, err := database.NewPostgresPool(ctx, databaseURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	authorRepository := postgres.NewAuthorRepository(db)
	bookRepository := postgres.NewBookRepository(db)

	file, err := os.Open(inputFile)
	if err != nil {
		panic(fmt.Errorf("failed to open csv: %w", err))
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// пропускаем заголовок
	_, err = reader.Read()
	if err != nil {
		panic(fmt.Errorf("failed to read header: %w", err))
	}

	authors, err := authorRepository.GetAll(ctx)
	if err != nil {
		panic(err)
	}

	authorIDs := make(map[string]int64)

	for _, author := range authors {
		authorIDs[strings.ToLower(author.Name)] = author.ID
	}

	imported := 0

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
			author, err := authorRepository.Create(
				ctx,
				domain.CreateAuthorRequest{
					Name: authorName,
				},
			)

			if err != nil {
				panic(fmt.Errorf(
					"failed to create author %q: %w",
					authorName,
					err,
				))
			}

			authorID = author.ID
			authorIDs[authorKey] = authorID

			fmt.Printf("created author: %s\n", authorName)
		}

		_, err = bookRepository.Create(
			ctx,
			domain.CreateBookRequest{
				AuthorID:    authorID,
				Title:       title,
				Description: description,
			},
		)

		if err != nil {
			panic(fmt.Errorf(
				"failed to create book %q: %w",
				title,
				err,
			))
		}

		imported++

		fmt.Printf("[%d] imported: %s — %s\n",
			imported,
			title,
			authorName,
		)
	}

	fmt.Printf("\nImport completed: %d books\n", imported)
}