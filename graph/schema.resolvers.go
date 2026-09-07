package graph

import (
	"context"
	"errors"
	"strconv"

	"book-catalog-api/graph/model"
	"book-catalog-api/internal/apperror"
	"book-catalog-api/internal/domain"
)

func (r *queryResolver) Books(ctx context.Context) ([]*model.Book, error) {
	books, err := r.BookRepository.GetAll(ctx, domain.BookFilter{})
	if err != nil {
		return nil, err
	}

	result := make([]*model.Book, 0, len(books))

	for _, book := range books {
		result = append(result, toGraphQLBook(book))
	}

	return result, nil
}

func (r *queryResolver) Book(ctx context.Context, id string) (*model.Book, error) {
	bookID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || bookID <= 0 {
		return nil, errors.New("invalid book id")
	}

	book, err := r.BookRepository.GetByID(ctx, bookID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return toGraphQLBook(*book), nil
}

func (r *queryResolver) Authors(ctx context.Context) ([]*model.Author, error) {
	authors, err := r.AuthorRepository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Author, 0, len(authors))

	for _, author := range authors {
		result = append(result, toGraphQLAuthor(author))
	}

	return result, nil
}

func (r *queryResolver) Author(ctx context.Context, id string) (*model.Author, error) {
	authorID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || authorID <= 0 {
		return nil, errors.New("invalid author id")
	}

	author, err := r.AuthorRepository.GetByID(ctx, authorID)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return toGraphQLAuthorWithBooks(*author), nil
}

func toGraphQLBook(book domain.Book) *model.Book {
	description := book.Description

	return &model.Book{
		ID:          strconv.FormatInt(book.ID, 10),
		Title:       book.Title,
		Description: &description,
		Author: &model.Author{
			ID:   strconv.FormatInt(book.Author.ID, 10),
			Name: book.Author.Name,
		},
	}
}

func toGraphQLAuthor(author domain.Author) *model.Author {
	return &model.Author{
		ID:   strconv.FormatInt(author.ID, 10),
		Name: author.Name,
	}
}

func toGraphQLAuthorWithBooks(author domain.AuthorWithBooks) *model.Author {
	return &model.Author{
		ID:   strconv.FormatInt(author.ID, 10),
		Name: author.Name,
	}
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct {
	*Resolver
}
