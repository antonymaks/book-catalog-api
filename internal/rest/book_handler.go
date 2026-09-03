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

type BookHandler struct {
	repository *postgres.BookRepository
}

func NewBookHandler(repository *postgres.BookRepository) *BookHandler {
	return &BookHandler{
		repository: repository,
	}
}

func (h *BookHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	filter := domain.BookFilter{
		Search: r.URL.Query().Get("search"),
		Author: r.URL.Query().Get("author"),
		Sort:   r.URL.Query().Get("sort"),
		Order:  r.URL.Query().Get("order"),
	}

	authorIDStr := r.URL.Query().Get("author_id")

	if authorIDStr != "" {
		authorID, err := strconv.ParseInt(authorIDStr, 10, 64)
		if err != nil || authorID <= 0 {
			http.Error(
				w,
				`{"error":"invalid author_id"}`,
				http.StatusBadRequest,
			)
			return
		}

		filter.AuthorID = authorID
	}

	limitStr := r.URL.Query().Get("limit")

	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			http.Error(
				w,
				`{"error":"invalid limit"}`,
				http.StatusBadRequest,
			)
			return
		}

		if limit > 100 {
			limit = 100
		}

		filter.Limit = limit
	}

	offsetStr := r.URL.Query().Get("offset")

	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			http.Error(
				w,
				`{"error":"invalid offset"}`,
				http.StatusBadRequest,
			)
			return
		}

		filter.Offset = offset
	}

	if filter.Sort != "" &&
		filter.Sort != "id" &&
		filter.Sort != "title" &&
		filter.Sort != "author" {

		http.Error(
			w,
			`{"error":"invalid sort field"}`,
			http.StatusBadRequest,
		)
		return
	}

	if filter.Order != "" &&
		filter.Order != "asc" &&
		filter.Order != "desc" {

		http.Error(
			w,
			`{"error":"invalid order"}`,
			http.StatusBadRequest,
		)
		return
	}

	books, err := h.repository.GetAll(r.Context(), filter)
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to get books"}`,
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(books); err != nil {
		http.Error(
			w,
			`{"error":"failed to encode response"}`,
			http.StatusInternalServerError,
		)
	}
}

func (h *BookHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	book, err := h.repository.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			writeError(w, http.StatusNotFound, "book not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, book)
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateBookRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	if input.AuthorID <= 0 {
		writeError(w, http.StatusBadRequest, "author_id is required")
		return
	}

	book, err := h.repository.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, apperror.ErrInvalidReference) {
			writeError(w, http.StatusBadRequest, "author does not exist")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, book)
}

func (h *BookHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	var input domain.UpdateBookRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	if input.AuthorID <= 0 {
		writeError(w, http.StatusBadRequest, "author_id is required")
		return
	}

	book, err := h.repository.Update(r.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrNotFound):
			writeError(w, http.StatusNotFound, "book not found")

		case errors.Is(err, apperror.ErrInvalidReference):
			writeError(w, http.StatusBadRequest, "author does not exist")

		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}

		return
	}

	writeJSON(w, http.StatusOK, book)
}

func (h *BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	err = h.repository.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			writeError(w, http.StatusNotFound, "book not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}
