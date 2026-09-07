package graph

import "book-catalog-api/internal/repository/postgres"

type Resolver struct {
	BookRepository   *postgres.BookRepository
	AuthorRepository *postgres.AuthorRepository
}
