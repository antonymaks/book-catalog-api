package domain

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Используется только внутри приложения при авторизации.
// PasswordHash не должен попадать в обычные API-ответы.
type UserWithPassword struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         string
}

type CreateUserRequest struct {
	Username     string
	PasswordHash string
	Role         string
}

type AddToReadingListRequest struct {
	BookID int64 `json:"book_id"`
}
