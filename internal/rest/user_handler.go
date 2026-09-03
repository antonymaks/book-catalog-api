package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"book-catalog-api/internal/apperror"
	"book-catalog-api/internal/domain"
	"book-catalog-api/internal/repository/postgres"
)

type UserHandler struct {
	repository *postgres.UserRepository
}

func NewUserHandler(repository *postgres.UserRepository) *UserHandler {
	return &UserHandler{
		repository: repository,
	}
}

func (h *UserHandler) GetAll(
	w http.ResponseWriter,
	r *http.Request,
) {
	users, err := h.repository.GetAll(r.Context())
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to get users"}`,
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(
			w,
			`{"error":"invalid user id"}`,
			http.StatusBadRequest,
		)
		return
	}

	user, err := h.repository.GetByID(r.Context(), id)
	if err != nil {
		http.Error(
			w,
			`{"error":"user not found"}`,
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetReadingList(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || userID <= 0 {
		http.Error(
			w,
			`{"error":"invalid user id"}`,
			http.StatusBadRequest,
		)
		return
	}

	books, err := h.repository.GetReadingList(r.Context(), userID)
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to get reading list"}`,
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (h *UserHandler) AddToReadingList(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || userID <= 0 {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid user id",
		)
		return
	}

	var input domain.AddToReadingListRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	if input.BookID <= 0 {
		writeError(
			w,
			http.StatusBadRequest,
			"book_id is required",
		)
		return
	}

	err = h.repository.AddToReadingList(
		r.Context(),
		userID,
		input.BookID,
	)

	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrConflict):
			writeError(
				w,
				http.StatusConflict,
				"book is already in reading list",
			)

		case errors.Is(err, apperror.ErrInvalidReference):
			writeError(
				w,
				http.StatusNotFound,
				"user or book not found",
			)

		default:
			writeError(
				w,
				http.StatusInternalServerError,
				"internal server error",
			)
		}

		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		map[string]string{
			"message": "book added to reading list",
		},
	)
}

func (h *UserHandler) RemoveFromReadingList(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || userID <= 0 {
		http.Error(
			w,
			`{"error":"invalid user id"}`,
			http.StatusBadRequest,
		)
		return
	}

	bookID, err := strconv.ParseInt(
		r.PathValue("book_id"),
		10,
		64,
	)
	if err != nil || bookID <= 0 {
		http.Error(
			w,
			`{"error":"invalid book id"}`,
			http.StatusBadRequest,
		)
		return
	}

	err = h.repository.RemoveFromReadingList(
		r.Context(),
		userID,
		bookID,
	)

	if err != nil {
		http.Error(
			w,
			`{"error":"book not found in reading list"}`,
			http.StatusNotFound,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
