package mongo

import (
	"context"
	"errors"
	"fmt"

	"book-catalog-api/internal/apperror"
	"book-catalog-api/internal/domain"

	"go.mongodb.org/mongo-driver/v2/bson"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type BookRepository struct {
	db *mongodriver.Database
}

type bookDocument struct {
	ID          int64  `bson:"id"`
	AuthorID    int64  `bson:"author_id"`
	Title       string `bson:"title"`
	Description string `bson:"description"`
}

type bookWithAuthorDocument struct {
	ID          int64  `bson:"id"`
	Title       string `bson:"title"`
	Description string `bson:"description"`
	Author      struct {
		ID   int64  `bson:"id"`
		Name string `bson:"name"`
	} `bson:"author"`
}

func NewBookRepository(db *mongodriver.Database) *BookRepository {
	return &BookRepository{
		db: db,
	}
}

func (r *BookRepository) GetAll(
	ctx context.Context,
	filter domain.BookFilter,
) ([]domain.Book, error) {

	pipeline := mongodriver.Pipeline{
		{
			{"$lookup", bson.D{
				{"from", "authors"},
				{"localField", "author_id"},
				{"foreignField", "id"},
				{"as", "author"},
			}},
		},
		{
			{"$unwind", "$author"},
		},
	}

	match := bson.D{}

	if filter.Search != "" {
		regex := bson.Regex{
			Pattern: filter.Search,
			Options: "i",
		}

		match = append(match,
			bson.E{
				Key: "$or",
				Value: bson.A{
					bson.D{{"title", regex}},
					bson.D{{"description", regex}},
				},
			},
		)
	}

	if filter.Author != "" {
		match = append(match,
			bson.E{
				Key: "author.name",
				Value: bson.Regex{
					Pattern: filter.Author,
					Options: "i",
				},
			},
		)
	}

	if filter.AuthorID > 0 {
		match = append(match,
			bson.E{
				Key:   "author.id",
				Value: filter.AuthorID,
			},
		)
	}

	if len(match) > 0 {
		pipeline = append(
			pipeline,
			bson.D{{"$match", match}},
		)
	}

	sortField := "id"

	switch filter.Sort {
	case "title":
		sortField = "title"
	case "author":
		sortField = "author.name"
	case "id":
		sortField = "id"
	}

	sortOrder := 1

	if filter.Order == "desc" {
		sortOrder = -1
	}

	pipeline = append(
		pipeline,
		bson.D{
			{"$sort", bson.D{
				{sortField, sortOrder},
			}},
		},
	)

	if filter.Offset > 0 {
		pipeline = append(
			pipeline,
			bson.D{{"$skip", int64(filter.Offset)}},
		)
	}

	if filter.Limit > 0 {
		pipeline = append(
			pipeline,
			bson.D{{"$limit", int64(filter.Limit)}},
		)
	}

	cursor, err := r.db.
		Collection("books").
		Aggregate(ctx, pipeline)

	if err != nil {
		return nil, fmt.Errorf("failed to get books: %w", err)
	}

	defer cursor.Close(ctx)

	var documents []bookWithAuthorDocument

	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("failed to decode books: %w", err)
	}

	books := make([]domain.Book, 0, len(documents))

	for _, document := range documents {
		books = append(books, domain.Book{
			ID:          document.ID,
			Title:       document.Title,
			Description: document.Description,
			Author: domain.Author{
				ID:   document.Author.ID,
				Name: document.Author.Name,
			},
		})
	}

	return books, nil
}

func (r *BookRepository) GetByID(
	ctx context.Context,
	id int64,
) (*domain.Book, error) {

	pipeline := mongodriver.Pipeline{
		{
			{"$match", bson.D{
				{"id", id},
			}},
		},
		{
			{"$lookup", bson.D{
				{"from", "authors"},
				{"localField", "author_id"},
				{"foreignField", "id"},
				{"as", "author"},
			}},
		},
		{
			{"$unwind", "$author"},
		},
	}

	cursor, err := r.db.
		Collection("books").
		Aggregate(ctx, pipeline)

	if err != nil {
		return nil, fmt.Errorf("failed to get book by id: %w", err)
	}

	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		if err := cursor.Err(); err != nil {
			return nil, fmt.Errorf("failed to get book by id: %w", err)
		}

		return nil, apperror.ErrNotFound
	}

	var document bookWithAuthorDocument

	if err := cursor.Decode(&document); err != nil {
		return nil, fmt.Errorf("failed to decode book: %w", err)
	}

	book := &domain.Book{
		ID:          document.ID,
		Title:       document.Title,
		Description: document.Description,
		Author: domain.Author{
			ID:   document.Author.ID,
			Name: document.Author.Name,
		},
	}

	return book, nil
}

func (r *BookRepository) Create(
	ctx context.Context,
	input domain.CreateBookRequest,
) (*domain.Book, error) {

	authors := r.db.Collection("authors")

	err := authors.FindOne(
		ctx,
		bson.D{{"id", input.AuthorID}},
	).Err()

	if errors.Is(err, mongodriver.ErrNoDocuments) {
		return nil, apperror.ErrInvalidReference
	}

	if err != nil {
		return nil, fmt.Errorf("failed to check author: %w", err)
	}

	id, err := r.nextID(ctx)

	if err != nil {
		return nil, err
	}

	document := bookDocument{
		ID:          id,
		AuthorID:    input.AuthorID,
		Title:       input.Title,
		Description: input.Description,
	}

	_, err = r.db.
		Collection("books").
		InsertOne(ctx, document)

	if err != nil {
		return nil, fmt.Errorf("failed to create book: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *BookRepository) Update(
	ctx context.Context,
	id int64,
	input domain.UpdateBookRequest,
) (*domain.Book, error) {

	authors := r.db.Collection("authors")

	err := authors.FindOne(
		ctx,
		bson.D{{"id", input.AuthorID}},
	).Err()

	if errors.Is(err, mongodriver.ErrNoDocuments) {
		return nil, apperror.ErrInvalidReference
	}

	if err != nil {
		return nil, fmt.Errorf("failed to check author: %w", err)
	}

	update := bson.D{
		{"$set", bson.D{
			{"author_id", input.AuthorID},
			{"title", input.Title},
			{"description", input.Description},
		}},
	}

	result, err := r.db.
		Collection("books").
		UpdateOne(
			ctx,
			bson.D{{"id", id}},
			update,
		)

	if err != nil {
		return nil, fmt.Errorf("failed to update book: %w", err)
	}

	if result.MatchedCount == 0 {
		return nil, apperror.ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *BookRepository) Delete(
	ctx context.Context,
	id int64,
) error {

	result, err := r.db.
		Collection("books").
		DeleteOne(
			ctx,
			bson.D{{"id", id}},
		)

	if err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}

	if result.DeletedCount == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

func (r *BookRepository) nextID(
	ctx context.Context,
) (int64, error) {

	var document bookDocument

	err := r.db.
		Collection("books").
		FindOne(
			ctx,
			bson.D{},
			options.FindOne().SetSort(
				bson.D{{"id", -1}},
			),
		).
		Decode(&document)

	if errors.Is(err, mongodriver.ErrNoDocuments) {
		return 1, nil
	}

	if err != nil {
		return 0, fmt.Errorf(
			"failed to generate book id: %w",
			err,
		)
	}

	return document.ID + 1, nil
}
