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

type UserRepository struct {
	db *mongodriver.Database
}

type userDocument struct {
	ID           int64  `bson:"id"`
	Username     string `bson:"username"`
	PasswordHash string `bson:"password_hash"`
	Role         string `bson:"role"`
}

type readingListDocument struct {
	UserID int64 `bson:"user_id"`
	BookID int64 `bson:"book_id"`
}

func NewUserRepository(db *mongodriver.Database) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetAll(
	ctx context.Context,
) ([]domain.User, error) {

	cursor, err := r.db.
		Collection("users").
		Find(ctx, bson.D{})

	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	defer cursor.Close(ctx)

	var documents []userDocument

	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	users := make([]domain.User, 0, len(documents))

	for _, document := range documents {
		users = append(users, domain.User{
			ID:       document.ID,
			Username: document.Username,
			Role:     document.Role,
		})
	}

	return users, nil
}

func (r *UserRepository) GetByID(
	ctx context.Context,
	id int64,
) (*domain.User, error) {

	var document userDocument

	err := r.db.
		Collection("users").
		FindOne(
			ctx,
			bson.D{{"id", id}},
		).
		Decode(&document)

	if errors.Is(err, mongodriver.ErrNoDocuments) {
		return nil, apperror.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &domain.User{
		ID:       document.ID,
		Username: document.Username,
		Role:     document.Role,
	}, nil
}

func (r *UserRepository) GetReadingList(
	ctx context.Context,
	userID int64,
) ([]domain.Book, error) {

	// Проверяем существование пользователя.
	err := r.db.
		Collection("users").
		FindOne(
			ctx,
			bson.D{{"id", userID}},
		).
		Err()

	if errors.Is(err, mongodriver.ErrNoDocuments) {
		return nil, apperror.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to check user: %w", err)
	}

	pipeline := mongodriver.Pipeline{
		{
			{"$match", bson.D{
				{"user_id", userID},
			}},
		},
		{
			{"$lookup", bson.D{
				{"from", "books"},
				{"localField", "book_id"},
				{"foreignField", "id"},
				{"as", "book"},
			}},
		},
		{
			{"$unwind", "$book"},
		},
		{
			{"$lookup", bson.D{
				{"from", "authors"},
				{"localField", "book.author_id"},
				{"foreignField", "id"},
				{"as", "author"},
			}},
		},
		{
			{"$unwind", "$author"},
		},
		{
			{"$sort", bson.D{
				{"book.id", 1},
			}},
		},
	}

	cursor, err := r.db.
		Collection("reading_list").
		Aggregate(ctx, pipeline)

	if err != nil {
		return nil, fmt.Errorf("failed to get reading list: %w", err)
	}

	defer cursor.Close(ctx)

	type readingListResult struct {
		Book struct {
			ID          int64  `bson:"id"`
			Title       string `bson:"title"`
			Description string `bson:"description"`
		} `bson:"book"`

		Author struct {
			ID   int64  `bson:"id"`
			Name string `bson:"name"`
		} `bson:"author"`
	}

	var documents []readingListResult

	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("failed to decode reading list: %w", err)
	}

	books := make([]domain.Book, 0, len(documents))

	for _, document := range documents {
		books = append(books, domain.Book{
			ID:          document.Book.ID,
			Title:       document.Book.Title,
			Description: document.Book.Description,
			Author: domain.Author{
				ID:   document.Author.ID,
				Name: document.Author.Name,
			},
		})
	}

	return books, nil
}

func (r *UserRepository) AddToReadingList(
	ctx context.Context,
	userID int64,
	bookID int64,
) error {

	// В PostgreSQL существование пользователя и книги
	// контролируется внешними ключами.
	// MongoDB таких связей не проверяет, поэтому делаем это сами.

	err := r.db.
		Collection("users").
		FindOne(
			ctx,
			bson.D{{"id", userID}},
		).
		Err()

	if errors.Is(err, mongodriver.ErrNoDocuments) {
		return apperror.ErrInvalidReference
	}

	if err != nil {
		return fmt.Errorf("failed to check user: %w", err)
	}

	err = r.db.
		Collection("books").
		FindOne(
			ctx,
			bson.D{{"id", bookID}},
		).
		Err()

	if errors.Is(err, mongodriver.ErrNoDocuments) {
		return apperror.ErrInvalidReference
	}

	if err != nil {
		return fmt.Errorf("failed to check book: %w", err)
	}

	// Проверяем, нет ли уже такой книги в списке.
	err = r.db.
		Collection("reading_list").
		FindOne(
			ctx,
			bson.D{
				{"user_id", userID},
				{"book_id", bookID},
			},
		).
		Err()

	if err == nil {
		return apperror.ErrConflict
	}

	if !errors.Is(err, mongodriver.ErrNoDocuments) {
		return fmt.Errorf(
			"failed to check reading list item: %w",
			err,
		)
	}

	document := readingListDocument{
		UserID: userID,
		BookID: bookID,
	}

	_, err = r.db.
		Collection("reading_list").
		InsertOne(ctx, document)

	if err != nil {
		return fmt.Errorf("failed to add book to reading list: %w", err)
	}

	return nil
}

func (r *UserRepository) RemoveFromReadingList(
	ctx context.Context,
	userID int64,
	bookID int64,
) error {

	result, err := r.db.
		Collection("reading_list").
		DeleteOne(
			ctx,
			bson.D{
				{"user_id", userID},
				{"book_id", bookID},
			},
		)

	if err != nil {
		return fmt.Errorf(
			"failed to remove book from reading list: %w",
			err,
		)
	}

	if result.DeletedCount == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

func (r *UserRepository) GetByUsername(
	ctx context.Context,
	username string,
) (*domain.UserWithPassword, error) {

	var document userDocument

	err := r.db.
		Collection("users").
		FindOne(
			ctx,
			bson.D{
				{"username", username},
			},
		).
		Decode(&document)

	if errors.Is(err, mongodriver.ErrNoDocuments) {
		return nil, apperror.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get user by username: %w",
			err,
		)
	}

	return &domain.UserWithPassword{
		ID:           document.ID,
		Username:     document.Username,
		PasswordHash: document.PasswordHash,
		Role:         document.Role,
	}, nil
}

func (r *UserRepository) Create(
	ctx context.Context,
	input domain.CreateUserRequest,
) (*domain.User, error) {

	id, err := r.nextUserID(ctx)
	if err != nil {
		return nil, err
	}

	document := userDocument{
		ID:           id,
		Username:     input.Username,
		PasswordHash: input.PasswordHash,
		Role:         input.Role,
	}

	_, err = r.db.
		Collection("users").
		InsertOne(
			ctx,
			document,
		)

	if mongodriver.IsDuplicateKeyError(err) {
		return nil, apperror.ErrConflict
	}

	if err != nil {
		return nil, fmt.Errorf(
			"failed to create user: %w",
			err,
		)
	}

	return &domain.User{
		ID:       document.ID,
		Username: document.Username,
		Role:     document.Role,
	}, nil
}

func (r *UserRepository) nextUserID(
	ctx context.Context,
) (int64, error) {

	var document userDocument

	err := r.db.
		Collection("users").
		FindOne(
			ctx,
			bson.D{},
			options.FindOne().
				SetSort(
					bson.D{{"id", -1}},
				),
		).
		Decode(&document)

	if errors.Is(err, mongodriver.ErrNoDocuments) {
		return 1, nil
	}

	if err != nil {
		return 0, fmt.Errorf(
			"failed to generate user id: %w",
			err,
		)
	}

	return document.ID + 1, nil
}
