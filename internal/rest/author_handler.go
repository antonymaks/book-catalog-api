package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"book-catalog-api/internal/apperror"
	"book-catalog-api/internal/domain"
	"book-catalog-api/internal/repository/postgres"
)

type AuthorHandler struct {
	repository *postgres.AuthorRepository
}

func NewAuthorHandler(
	repository *postgres.AuthorRepository,
) *AuthorHandler {
	return &AuthorHandler{
		repository: repository,
	}
}

func (h *AuthorHandler) GetAll(
	w http.ResponseWriter,
	r *http.Request,
) {
	authors, err := h.repository.GetAll(r.Context())
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to get authors"}`,
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authors)
}

func (h *AuthorHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(
			w,
			`{"error":"invalid author id"}`,
			http.StatusBadRequest,
		)
		return
	}

	author, err := h.repository.GetByID(r.Context(), id)
	if err != nil {
		http.Error(
			w,
			`{"error":"author not found"}`,
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(author)
}

func (h *AuthorHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var input domain.CreateAuthorRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(
			w,
			`{"error":"invalid request body"}`,
			http.StatusBadRequest,
		)
		return
	}

	input.Name = strings.TrimSpace(input.Name)

	if input.Name == "" {
		http.Error(
			w,
			`{"error":"name is required"}`,
			http.StatusBadRequest,
		)
		return
	}

	author, err := h.repository.Create(r.Context(), input)
	if err != nil {
		http.Error(
			w,
			`{"error":"failed to create author"}`,
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(author)
}

func (h *AuthorHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(
			w,
			`{"error":"invalid author id"}`,
			http.StatusBadRequest,
		)
		return
	}

	var input domain.UpdateAuthorRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(
			w,
			`{"error":"invalid request body"}`,
			http.StatusBadRequest,
		)
		return
	}

	input.Name = strings.TrimSpace(input.Name)

	if input.Name == "" {
		http.Error(
			w,
			`{"error":"name is required"}`,
			http.StatusBadRequest,
		)
		return
	}

	author, err := h.repository.Update(r.Context(), id, input)
	if err != nil {
		http.Error(
			w,
			`{"error":"author not found"}`,
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(author)
}

func (h *AuthorHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid author id")
		return
	}

	err = h.repository.Delete(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, apperror.ErrNotFound):
			writeError(w, http.StatusNotFound, "author not found")

		case errors.Is(err, apperror.ErrConflict):
			writeError(
				w,
				http.StatusConflict,
				"author cannot be deleted because they have books",
			)

		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}

		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}
