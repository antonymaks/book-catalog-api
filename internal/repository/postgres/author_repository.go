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

type AuthorRepository struct {
	db *pgxpool.Pool
}

func NewAuthorRepository(db *pgxpool.Pool) *AuthorRepository {
	return &AuthorRepository{db: db}
}

func (r *AuthorRepository) GetAll(
	ctx context.Context,
) ([]domain.Author, error) {
	query := `
		SELECT id, name
		FROM authors
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get authors: %w", err)
	}
	defer rows.Close()

	authors := make([]domain.Author, 0)

	for rows.Next() {
		var author domain.Author

		if err := rows.Scan(
			&author.ID,
			&author.Name,
		); err != nil {
			return nil, fmt.Errorf("failed to scan author: %w", err)
		}

		authors = append(authors, author)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return authors, nil
}

func (r *AuthorRepository) GetByID(
	ctx context.Context,
	id int64,
) (*domain.AuthorWithBooks, error) {
	query := `
		SELECT id, name
		FROM authors
		WHERE id = $1
	`

	var author domain.AuthorWithBooks

	err := r.db.QueryRow(ctx, query, id).Scan(
		&author.ID,
		&author.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	booksQuery := `
		SELECT id, title, description
		FROM books
		WHERE author_id = $1
		ORDER BY title
	`

	rows, err := r.db.Query(ctx, booksQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get author books: %w", err)
	}
	defer rows.Close()

	author.Books = make([]domain.Book, 0)

	for rows.Next() {
		var book domain.Book

		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan book: %w", err)
		}

		book.Author = domain.Author{
			ID:   author.ID,
			Name: author.Name,
		}

		author.Books = append(author.Books, book)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return &author, nil
}

func (r *AuthorRepository) Create(
	ctx context.Context,
	input domain.CreateAuthorRequest,
) (*domain.Author, error) {
	query := `
		INSERT INTO authors (name)
		VALUES ($1)
		RETURNING id, name
	`

	var author domain.Author

	err := r.db.QueryRow(
		ctx,
		query,
		input.Name,
	).Scan(
		&author.ID,
		&author.Name,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create author: %w", err)
	}

	return &author, nil
}

func (r *AuthorRepository) Update(
	ctx context.Context,
	id int64,
	input domain.UpdateAuthorRequest,
) (*domain.Author, error) {
	query := `
		UPDATE authors
		SET name = $1
		WHERE id = $2
		RETURNING id, name
	`

	var author domain.Author

	err := r.db.QueryRow(
		ctx,
		query,
		input.Name,
		id,
	).Scan(
		&author.ID,
		&author.Name,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update author: %w", err)
	}

	return &author, nil
}

func (r *AuthorRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	query := `
		DELETE FROM authors
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				return apperror.ErrConflict
			}
		}

		return fmt.Errorf("failed to delete author: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}
