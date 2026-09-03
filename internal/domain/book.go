package domain

type Author struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Book struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Author      Author `json:"author"`
}

type CreateBookRequest struct {
	AuthorID    int64  `json:"author_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateBookRequest struct {
	AuthorID    int64  `json:"author_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type BookFilter struct {
	Search   string
	Author   string
	AuthorID int64

	Sort   string
	Order  string
	Limit  int
	Offset int
}
