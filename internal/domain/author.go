package domain

type CreateAuthorRequest struct {
	Name string `json:"name"`
}

type UpdateAuthorRequest struct {
	Name string `json:"name"`
}

type AuthorWithBooks struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Books []Book `json:"books"`
}
