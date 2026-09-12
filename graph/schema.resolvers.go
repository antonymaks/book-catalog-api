package graph

import (
	"context"
	"errors"
	"strconv"

	"book-catalog-api/graph/model"
	"book-catalog-api/internal/apperror"
	"book-catalog-api/internal/auth"
	"book-catalog-api/internal/domain"
)

// CreateBook is the resolver for the createBook field.
func (r *mutationResolver) CreateBook(
	ctx context.Context,
	input model.CreateBookInput,
) (*model.Book, error) {

	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	authorID, err := strconv.ParseInt(
		input.AuthorID,
		10,
		64,
	)
	if err != nil || authorID <= 0 {
		return nil, errors.New("invalid author id")
	}

	if input.Title == "" {
		return nil, errors.New("title is required")
	}

	description := ""

	if input.Description != nil {
		description = *input.Description
	}

	book, err := r.BookRepository.Create(
		ctx,
		domain.CreateBookRequest{
			AuthorID:    authorID,
			Title:       input.Title,
			Description: description,
		},
	)

	if err != nil {
		switch {
		case errors.Is(
			err,
			apperror.ErrInvalidReference,
		):
			return nil, errors.New("author not found")

		default:
			return nil, err
		}
	}

	return toGraphQLBook(*book), nil
}

// UpdateBook is the resolver for the updateBook field.
func (r *mutationResolver) UpdateBook(
	ctx context.Context,
	id string,
	input model.UpdateBookInput,
) (*model.Book, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	bookID, err := strconv.ParseInt(
		id,
		10,
		64,
	)
	if err != nil || bookID <= 0 {
		return nil, errors.New("invalid book id")
	}

	authorID, err := strconv.ParseInt(
		input.AuthorID,
		10,
		64,
	)
	if err != nil || authorID <= 0 {
		return nil, errors.New("invalid author id")
	}

	if input.Title == "" {
		return nil, errors.New("title is required")
	}

	description := ""

	if input.Description != nil {
		description = *input.Description
	}

	book, err := r.BookRepository.Update(
		ctx,
		bookID,
		domain.UpdateBookRequest{
			AuthorID:    authorID,
			Title:       input.Title,
			Description: description,
		},
	)

	if err != nil {
		switch {
		case errors.Is(
			err,
			apperror.ErrNotFound,
		):
			return nil, errors.New("book not found")

		case errors.Is(
			err,
			apperror.ErrInvalidReference,
		):
			return nil, errors.New("author not found")

		default:
			return nil, err
		}
	}

	return toGraphQLBook(*book), nil
}

// DeleteBook is the resolver for the deleteBook field.
func (r *mutationResolver) DeleteBook(
	ctx context.Context,
	id string,
) (bool, error) {
	if err := requireAdmin(ctx); err != nil {
		return false, err
	}

	bookID, err := strconv.ParseInt(
		id,
		10,
		64,
	)
	if err != nil || bookID <= 0 {
		return false, errors.New("invalid book id")
	}

	err = r.BookRepository.Delete(
		ctx,
		bookID,
	)
	if err != nil {
		if errors.Is(
			err,
			apperror.ErrNotFound,
		) {
			return false, errors.New("book not found")
		}

		return false, err
	}

	return true, nil
}

// CreateAuthor is the resolver for the createAuthor field.
func (r *mutationResolver) CreateAuthor(
	ctx context.Context,
	input model.CreateAuthorInput,
) (*model.Author, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	if input.Name == "" {
		return nil, errors.New("name is required")
	}

	author, err := r.AuthorRepository.Create(
		ctx,
		domain.CreateAuthorRequest{
			Name: input.Name,
		},
	)
	if err != nil {
		return nil, err
	}

	return toGraphQLAuthor(*author), nil
}

// UpdateAuthor is the resolver for the updateAuthor field.
func (r *mutationResolver) UpdateAuthor(
	ctx context.Context,
	id string,
	input model.UpdateAuthorInput,
) (*model.Author, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}

	authorID, err := strconv.ParseInt(
		id,
		10,
		64,
	)
	if err != nil || authorID <= 0 {
		return nil, errors.New("invalid author id")
	}

	if input.Name == "" {
		return nil, errors.New("name is required")
	}

	author, err := r.AuthorRepository.Update(
		ctx,
		authorID,
		domain.UpdateAuthorRequest{
			Name: input.Name,
		},
	)

	if err != nil {
		if errors.Is(
			err,
			apperror.ErrNotFound,
		) {
			return nil, errors.New("author not found")
		}

		return nil, err
	}

	return toGraphQLAuthor(*author), nil
}

// DeleteAuthor is the resolver for the deleteAuthor field.
func (r *mutationResolver) DeleteAuthor(
	ctx context.Context,
	id string,
) (bool, error) {
	if err := requireAdmin(ctx); err != nil {
		return false, err
	}

	authorID, err := strconv.ParseInt(
		id,
		10,
		64,
	)
	if err != nil || authorID <= 0 {
		return false, errors.New("invalid author id")
	}

	err = r.AuthorRepository.Delete(
		ctx,
		authorID,
	)

	if err != nil {
		switch {
		case errors.Is(
			err,
			apperror.ErrNotFound,
		):
			return false, errors.New("author not found")

		case errors.Is(
			err,
			apperror.ErrConflict,
		):
			return false, errors.New(
				"author cannot be deleted because they have books",
			)

		default:
			return false, err
		}
	}

	return true, nil
}

// AddBookToReadingList is the resolver for the addBookToReadingList field.
func (r *mutationResolver) AddBookToReadingList(
	ctx context.Context,
	bookID string,
) (bool, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return false, errors.New(
			"authentication required",
		)
	}

	parsedBookID, err := strconv.ParseInt(
		bookID,
		10,
		64,
	)
	if err != nil || parsedBookID <= 0 {
		return false, errors.New("invalid book id")
	}

	err = r.UserRepository.AddToReadingList(
		ctx,
		userID,
		parsedBookID,
	)

	if err != nil {
		switch {
		case errors.Is(
			err,
			apperror.ErrConflict,
		):
			return false, errors.New(
				"book is already in reading list",
			)

		case errors.Is(
			err,
			apperror.ErrInvalidReference,
		):
			return false, errors.New(
				"user or book not found",
			)

		default:
			return false, err
		}
	}

	return true, nil
}

// RemoveBookFromReadingList is the resolver for the removeBookFromReadingList field.
func (r *mutationResolver) RemoveBookFromReadingList(
	ctx context.Context,
	bookID string,
) (bool, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return false, errors.New(
			"authentication required",
		)
	}

	parsedBookID, err := strconv.ParseInt(
		bookID,
		10,
		64,
	)
	if err != nil || parsedBookID <= 0 {
		return false, errors.New("invalid book id")
	}

	err = r.UserRepository.RemoveFromReadingList(
		ctx,
		userID,
		parsedBookID,
	)

	if err != nil {
		if errors.Is(
			err,
			apperror.ErrNotFound,
		) {
			return false, errors.New(
				"book is not in reading list",
			)
		}

		return false, err
	}

	return true, nil
}

// Books is the resolver for the books field.
func (r *queryResolver) Books(
	ctx context.Context,
	filter *model.BookFilter,
) ([]*model.Book, error) {
	domainFilter := domain.BookFilter{}

	if filter != nil {
		if filter.Search != nil {
			domainFilter.Search = *filter.Search
		}

		if filter.Author != nil {
			domainFilter.Author = *filter.Author
		}

		if filter.AuthorID != nil {
			authorID, err := strconv.ParseInt(
				*filter.AuthorID,
				10,
				64,
			)
			if err != nil || authorID <= 0 {
				return nil, errors.New(
					"invalid author id",
				)
			}

			domainFilter.AuthorID = authorID
		}

		if filter.Sort != nil {
			domainFilter.Sort = *filter.Sort
		}

		if filter.Order != nil {
			domainFilter.Order = *filter.Order
		}

		if filter.Limit != nil {
			domainFilter.Limit = int(
				*filter.Limit,
			)
		}

		if filter.Offset != nil {
			domainFilter.Offset = int(
				*filter.Offset,
			)
		}
	}

	books, err := r.BookRepository.GetAll(
		ctx,
		domainFilter,
	)
	if err != nil {
		return nil, err
	}

	result := make(
		[]*model.Book,
		0,
		len(books),
	)

	for _, book := range books {
		result = append(
			result,
			toGraphQLBook(book),
		)
	}

	return result, nil
}

// Book is the resolver for the book field.
func (r *queryResolver) Book(
	ctx context.Context,
	id string,
) (*model.Book, error) {
	bookID, err := strconv.ParseInt(
		id,
		10,
		64,
	)
	if err != nil || bookID <= 0 {
		return nil, errors.New("invalid book id")
	}

	book, err := r.BookRepository.GetByID(
		ctx,
		bookID,
	)

	if err != nil {
		if errors.Is(
			err,
			apperror.ErrNotFound,
		) {
			return nil, nil
		}

		return nil, err
	}

	return toGraphQLBook(*book), nil
}

// Authors is the resolver for the authors field.
func (r *queryResolver) Authors(
	ctx context.Context,
) ([]*model.Author, error) {
	authors, err := r.AuthorRepository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make(
		[]*model.Author,
		0,
		len(authors),
	)

	for _, author := range authors {
		result = append(
			result,
			toGraphQLAuthor(author),
		)
	}

	return result, nil
}

// Author is the resolver for the author field.
func (r *queryResolver) Author(
	ctx context.Context,
	id string,
) (*model.Author, error) {
	authorID, err := strconv.ParseInt(
		id,
		10,
		64,
	)
	if err != nil || authorID <= 0 {
		return nil, errors.New("invalid author id")
	}

	author, err := r.AuthorRepository.GetByID(
		ctx,
		authorID,
	)

	if err != nil {
		if errors.Is(
			err,
			apperror.ErrNotFound,
		) {
			return nil, nil
		}

		return nil, err
	}

	return toGraphQLAuthorWithBooks(
		*author,
	), nil
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(
	ctx context.Context,
) ([]*model.User, error) {
	users, err := r.UserRepository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make(
		[]*model.User,
		0,
		len(users),
	)

	for _, user := range users {
		result = append(
			result,
			toGraphQLUser(user),
		)
	}

	return result, nil
}

// User is the resolver for the user field.
func (r *queryResolver) User(
	ctx context.Context,
	id string,
) (*model.User, error) {
	userID, err := strconv.ParseInt(
		id,
		10,
		64,
	)
	if err != nil || userID <= 0 {
		return nil, errors.New("invalid user id")
	}

	user, err := r.UserRepository.GetByID(
		ctx,
		userID,
	)

	if err != nil {
		if errors.Is(
			err,
			apperror.ErrNotFound,
		) {
			return nil, nil
		}

		return nil, err
	}

	books, err := r.UserRepository.GetReadingList(
		ctx,
		userID,
	)
	if err != nil {
		return nil, err
	}

	graphQLUser := toGraphQLUser(*user)

	graphQLUser.ReadingList = make(
		[]*model.Book,
		0,
		len(books),
	)

	for _, book := range books {
		graphQLUser.ReadingList = append(
			graphQLUser.ReadingList,
			toGraphQLBook(book),
		)
	}

	return graphQLUser, nil
}

// ReadingList is the resolver for the readingList field.
func (r *queryResolver) ReadingList(
	ctx context.Context,
) ([]*model.Book, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, errors.New(
			"authentication required",
		)
	}

	books, err := r.UserRepository.GetReadingList(
		ctx,
		userID,
	)
	if err != nil {
		return nil, err
	}

	result := make(
		[]*model.Book,
		0,
		len(books),
	)

	for _, book := range books {
		result = append(
			result,
			toGraphQLBook(book),
		)
	}

	return result, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

func requireAdmin(ctx context.Context) error {
	role, ok := auth.RoleFromContext(ctx)

	if !ok {
		return errors.New("authentication required")
	}

	if role != "admin" {
		return errors.New("admin access required")
	}

	return nil
}

type (
	mutationResolver struct {
		*Resolver
	}

	queryResolver struct {
		*Resolver
	}
)
