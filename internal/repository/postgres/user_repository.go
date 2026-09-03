package postgres

import (
	"context"
	"errors"
	"fmt"

	"book-catalog-api/internal/apperror"
	"book-catalog-api/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetAll(
	ctx context.Context,
) ([]domain.User, error) {
	query := `
		SELECT id, username, role
		FROM users
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0)

	for rows.Next() {
		var user domain.User

		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Role,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) GetByID(
	ctx context.Context,
	id int64,
) (*domain.User, error) {
	query := `
		SELECT id, username, role
		FROM users
		WHERE id = $1
	`

	var user domain.User

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetReadingList(
	ctx context.Context,
	userID int64,
) ([]domain.Book, error) {
	query := `
		SELECT
			b.id,
			b.title,
			b.description,
			a.id,
			a.name
		FROM reading_list rl
		JOIN books b ON rl.book_id = b.id
		JOIN authors a ON b.author_id = a.id
		WHERE rl.user_id = $1
		ORDER BY b.title
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reading list: %w", err)
	}
	defer rows.Close()

	books := make([]domain.Book, 0)

	for rows.Next() {
		var book domain.Book

		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Description,
			&book.Author.ID,
			&book.Author.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan book: %w", err)
		}

		books = append(books, book)
	}

	return books, nil
}

func (r *UserRepository) AddToReadingList(
	ctx context.Context,
	userID int64,
	bookID int64,
) error {
	query := `
		INSERT INTO reading_list (user_id, book_id)
		VALUES ($1, $2)
	`

	_, err := r.db.Exec(ctx, query, userID, bookID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return apperror.ErrConflict

			case "23503":
				return apperror.ErrInvalidReference
			}
		}

		return fmt.Errorf("failed to add book to reading list: %w", err)
	}

	return nil
}

func (r *UserRepository) RemoveFromReadingList(
	ctx context.Context,
	userID int64,
	bookID int64,
) error {
	query := `
		DELETE FROM reading_list
		WHERE user_id = $1
		AND book_id = $2
	`

	result, err := r.db.Exec(ctx, query, userID, bookID)
	if err != nil {
		return fmt.Errorf("failed to remove book from reading list: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}
