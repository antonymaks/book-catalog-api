package graph

import "book-catalog-api/internal/repository"

type Resolver struct {
	BookRepository   repository.BookRepository
	AuthorRepository repository.AuthorRepository
	UserRepository   repository.UserRepository
}
