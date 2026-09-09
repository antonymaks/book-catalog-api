package graph

import (
	"strconv"

	"book-catalog-api/graph/model"
	"book-catalog-api/internal/domain"
)

func toGraphQLBook(book domain.Book) *model.Book {
	description := book.Description

	return &model.Book{
		ID:          strconv.FormatInt(book.ID, 10),
		Title:       book.Title,
		Description: &description,
		Author: &model.Author{
			ID:    strconv.FormatInt(book.Author.ID, 10),
			Name:  book.Author.Name,
			Books: []*model.Book{},
		},
	}
}

func toGraphQLAuthor(author domain.Author) *model.Author {
	return &model.Author{
		ID:    strconv.FormatInt(author.ID, 10),
		Name:  author.Name,
		Books: []*model.Book{},
	}
}

func toGraphQLAuthorWithBooks(
	author domain.AuthorWithBooks,
) *model.Author {
	books := make([]*model.Book, 0, len(author.Books))

	for _, book := range author.Books {
		books = append(books, toGraphQLBook(book))
	}

	return &model.Author{
		ID:    strconv.FormatInt(author.ID, 10),
		Name:  author.Name,
		Books: books,
	}
}

func toGraphQLUser(user domain.User) *model.User {
	return &model.User{
		ID:          strconv.FormatInt(user.ID, 10),
		Username:    user.Username,
		Role:        user.Role,
		ReadingList: []*model.Book{},
	}
}
