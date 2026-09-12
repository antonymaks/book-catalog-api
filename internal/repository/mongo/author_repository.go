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

type AuthorRepository struct {
	db *mongodriver.Database
}

type authorDocument struct {
	ID   int64  `bson:"id"`
	Name string `bson:"name"`
}

func NewAuthorRepository(db *mongodriver.Database) *AuthorRepository {
	return &AuthorRepository{
		db: db,
	}
}

func (r *AuthorRepository) GetAll(
	ctx context.Context,
) ([]domain.Author, error) {

	cursor, err := r.db.
		Collection("authors").
		Find(
			ctx,
			bson.D{},
			options.Find().SetSort(
				bson.D{{"name", 1}},
			),
		)

	if err != nil {
		return nil, fmt.Errorf("failed to get authors: %w", err)
	}

	defer cursor.Close(ctx)

	var documents []authorDocument

	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("failed to decode authors: %w", err)
	}

	authors := make([]domain.Author, 0, len(documents))

	for _, document := range documents {
		authors = append(authors, domain.Author{
			ID:   document.ID,
			Name: document.Name,
		})
	}

	return authors, nil
}

func (r *AuthorRepository) GetByID(
	ctx context.Context,
	id int64,
) (*domain.AuthorWithBooks, error) {

	var document authorDocument

	err := r.db.
		Collection("authors").
		FindOne(
			ctx,
			bson.D{{"id", id}},
		).
		Decode(&document)

	if errors.Is(err, mongodriver.ErrNoDocuments) {
		return nil, apperror.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	author := &domain.AuthorWithBooks{
		ID:    document.ID,
		Name:  document.Name,
		Books: make([]domain.Book, 0),
	}

	cursor, err := r.db.
		Collection("books").
		Find(
			ctx,
			bson.D{{"author_id", id}},
			options.Find().SetSort(
				bson.D{{"title", 1}},
			),
		)

	if err != nil {
		return nil, fmt.Errorf("failed to get author books: %w", err)
	}

	defer cursor.Close(ctx)

	var books []bookDocument

	if err := cursor.All(ctx, &books); err != nil {
		return nil, fmt.Errorf("failed to decode author books: %w", err)
	}

	for _, book := range books {
		author.Books = append(author.Books, domain.Book{
			ID:          book.ID,
			Title:       book.Title,
			Description: book.Description,
			Author: domain.Author{
				ID:   document.ID,
				Name: document.Name,
			},
		})
	}

	return author, nil
}

func (r *AuthorRepository) Create(
	ctx context.Context,
	input domain.CreateAuthorRequest,
) (*domain.Author, error) {

	id, err := r.nextID(ctx)
	if err != nil {
		return nil, err
	}

	document := authorDocument{
		ID:   id,
		Name: input.Name,
	}

	_, err = r.db.
		Collection("authors").
		InsertOne(ctx, document)

	if err != nil {
		return nil, fmt.Errorf("failed to create author: %w", err)
	}

	return &domain.Author{
		ID:   document.ID,
		Name: document.Name,
	}, nil
}

func (r *AuthorRepository) Update(
	ctx context.Context,
	id int64,
	input domain.UpdateAuthorRequest,
) (*domain.Author, error) {

	result, err := r.db.
		Collection("authors").
		UpdateOne(
			ctx,
			bson.D{{"id", id}},
			bson.D{
				{"$set", bson.D{
					{"name", input.Name},
				}},
			},
		)

	if err != nil {
		return nil, fmt.Errorf("failed to update author: %w", err)
	}

	if result.MatchedCount == 0 {
		return nil, apperror.ErrNotFound
	}

	return &domain.Author{
		ID:   id,
		Name: input.Name,
	}, nil
}

func (r *AuthorRepository) Delete(
	ctx context.Context,
	id int64,
) error {

	// В PostgreSQL автора нельзя удалить,
	// если на него ссылаются книги.
	// В MongoDB такую проверку делаем вручную.
	count, err := r.db.
		Collection("books").
		CountDocuments(
			ctx,
			bson.D{{"author_id", id}},
		)

	if err != nil {
		return fmt.Errorf("failed to check author books: %w", err)
	}

	if count > 0 {
		return apperror.ErrConflict
	}

	result, err := r.db.
		Collection("authors").
		DeleteOne(
			ctx,
			bson.D{{"id", id}},
		)

	if err != nil {
		return fmt.Errorf("failed to delete author: %w", err)
	}

	if result.DeletedCount == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

func (r *AuthorRepository) nextID(
	ctx context.Context,
) (int64, error) {

	var document authorDocument

	err := r.db.
		Collection("authors").
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
			"failed to generate author id: %w",
			err,
		)
	}

	return document.ID + 1, nil
}
