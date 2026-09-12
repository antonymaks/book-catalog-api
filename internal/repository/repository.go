package repository

import (
	"context"

	"book-catalog-api/internal/domain"
)

type BookRepository interface {
	GetAll(
		ctx context.Context,
		filter domain.BookFilter,
	) ([]domain.Book, error)

	GetByID(
		ctx context.Context,
		id int64,
	) (*domain.Book, error)

	Create(
		ctx context.Context,
		input domain.CreateBookRequest,
	) (*domain.Book, error)

	Update(
		ctx context.Context,
		id int64,
		input domain.UpdateBookRequest,
	) (*domain.Book, error)

	Delete(
		ctx context.Context,
		id int64,
	) error
}

type AuthorRepository interface {
	GetAll(
		ctx context.Context,
	) ([]domain.Author, error)

	GetByID(
		ctx context.Context,
		id int64,
	) (*domain.AuthorWithBooks, error)

	Create(
		ctx context.Context,
		input domain.CreateAuthorRequest,
	) (*domain.Author, error)

	Update(
		ctx context.Context,
		id int64,
		input domain.UpdateAuthorRequest,
	) (*domain.Author, error)

	Delete(
		ctx context.Context,
		id int64,
	) error
}

type UserRepository interface {
	GetAll(
		ctx context.Context,
	) ([]domain.User, error)

	GetByID(
		ctx context.Context,
		id int64,
	) (*domain.User, error)

	GetReadingList(
		ctx context.Context,
		userID int64,
	) ([]domain.Book, error)

	AddToReadingList(
		ctx context.Context,
		userID int64,
		bookID int64,
	) error

	RemoveFromReadingList(
		ctx context.Context,
		userID int64,
		bookID int64,
	) error

	GetByUsername(
		ctx context.Context,
		username string,
	) (*domain.UserWithPassword, error)

	Create(
		ctx context.Context,
		input domain.CreateUserRequest,
	) (*domain.User, error)
}
