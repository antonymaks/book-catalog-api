package postgres

import (
	"context"
	"errors"
	"fmt"

	"book-catalog-api/internal/apperror"
	"book-catalog-api/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookRepository struct {
	db *pgxpool.Pool
}

func NewBookRepository(db *pgxpool.Pool) *BookRepository {
	return &BookRepository{
		db: db,
	}
}

func (r *BookRepository) GetAll(
	ctx context.Context,
	filter domain.BookFilter,
) ([]domain.Book, error) {
	query := `
		SELECT
			b.id,
			b.title,
			b.description,
			a.id,
			a.name
		FROM books b
		JOIN authors a ON b.author_id = a.id
		WHERE 1 = 1
	`

	args := make([]any, 0)

	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")

		query += fmt.Sprintf(`
			AND (
				b.title ILIKE $%d
				OR b.description ILIKE $%d
			)
		`, len(args), len(args))
	}

	if filter.Author != "" {
		args = append(args, "%"+filter.Author+"%")

		query += fmt.Sprintf(`
			AND a.name ILIKE $%d
		`, len(args))
	}

	if filter.AuthorID > 0 {
		args = append(args, filter.AuthorID)

		query += fmt.Sprintf(`
			AND a.id = $%d
		`, len(args))
	}

	switch filter.Sort {
	case "title":
		query += " ORDER BY b.title"
	case "author":
		query += " ORDER BY a.name"
	case "id":
		query += " ORDER BY b.id"
	default:
		query += " ORDER BY b.id"
	}

	if filter.Order == "desc" {
		query += " DESC"
	} else {
		query += " ASC"
	}

	if filter.Limit > 0 {
		args = append(args, filter.Limit)

		query += fmt.Sprintf(
			" LIMIT $%d",
			len(args),
		)
	}

	if filter.Offset > 0 {
		args = append(args, filter.Offset)

		query += fmt.Sprintf(
			" OFFSET $%d",
			len(args),
		)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get books: %w", err)
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return books, nil
}

func (r *BookRepository) GetByID(
	ctx context.Context,
	id int64,
) (*domain.Book, error) {
	query := `
		SELECT
			b.id,
			b.title,
			b.description,
			a.id,
			a.name
		FROM books b
		JOIN authors a ON b.author_id = a.id
		WHERE b.id = $1
	`

	var book domain.Book

	err := r.db.QueryRow(ctx, query, id).Scan(
		&book.ID,
		&book.Title,
		&book.Description,
		&book.Author.ID,
		&book.Author.Name,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}

		return nil, fmt.Errorf("failed to get book by id: %w", err)
	}

	return &book, nil
}

func (r *BookRepository) Create(
	ctx context.Context,
	input domain.CreateBookRequest,
) (*domain.Book, error) {
	query := `
		INSERT INTO books (author_id, title, description)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int64

	err := r.db.QueryRow(
		ctx,
		query,
		input.AuthorID,
		input.Title,
		input.Description,
	).Scan(&id)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				return nil, apperror.ErrInvalidReference
			}
		}

		return nil, fmt.Errorf("failed to create book: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *BookRepository) Update(
	ctx context.Context,
	id int64,
	input domain.UpdateBookRequest,
) (*domain.Book, error) {
	query := `
		UPDATE books
		SET author_id = $1,
			title = $2,
			description = $3
		WHERE id = $4
	`

	result, err := r.db.Exec(
		ctx,
		query,
		input.AuthorID,
		input.Title,
		input.Description,
		id,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				return nil, apperror.ErrInvalidReference
			}
		}

		return nil, fmt.Errorf("failed to update book: %w", err)
	}

	if result.RowsAffected() == 0 {
		return nil, apperror.ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *BookRepository) Delete(
	ctx context.Context,
	id int64,
) error {
	query := `
		DELETE FROM books
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperror.ErrNotFound
	}

	return nil
}
