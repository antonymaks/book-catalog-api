package domain

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type AddToReadingListRequest struct {
	BookID int64 `json:"book_id"`
}
