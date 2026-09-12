package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"book-catalog-api/internal/apperror"
	"book-catalog-api/internal/auth"
	"book-catalog-api/internal/repository"
)

type AuthHandler struct {
	authService *auth.Service
	users       repository.UserRepository
}

func NewAuthHandler(
	authService *auth.Service,
	users repository.UserRepository,
) *AuthHandler {

	return &AuthHandler{
		authService: authService,
		users:       users,
	}
}

func (h *AuthHandler) Register(
	w http.ResponseWriter,
	r *http.Request,
) {

	var input auth.RegisterInput

	if err := json.NewDecoder(
		r.Body,
	).Decode(&input); err != nil {

		writeError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)

		return
	}

	result, err :=
		h.authService.Register(
			r.Context(),
			input,
		)

	if err != nil {

		switch {

		case errors.Is(
			err,
			apperror.ErrConflict,
		):

			writeError(
				w,
				http.StatusConflict,
				"username already exists",
			)

		default:

			writeError(
				w,
				http.StatusBadRequest,
				err.Error(),
			)
		}

		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		result,
	)
}

func (h *AuthHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {

	var input auth.LoginInput

	if err := json.NewDecoder(
		r.Body,
	).Decode(&input); err != nil {

		writeError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)

		return
	}

	result, err :=
		h.authService.Login(
			r.Context(),
			input,
		)

	if err != nil {

		if errors.Is(
			err,
			apperror.ErrNotFound,
		) {

			writeError(
				w,
				http.StatusUnauthorized,
				"invalid username or password",
			)

			return
		}

		writeError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)

		return
	}

	writeJSON(
		w,
		http.StatusOK,
		result,
	)
}

func (h *AuthHandler) Me(
	w http.ResponseWriter,
	r *http.Request,
) {

	userID, ok :=
		auth.UserIDFromContext(
			r.Context(),
		)

	if !ok {
		writeError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)

		return
	}

	user, err :=
		h.users.GetByID(
			r.Context(),
			userID,
		)

	if err != nil {
		writeError(
			w,
			http.StatusNotFound,
			"user not found",
		)

		return
	}

	writeJSON(
		w,
		http.StatusOK,
		user,
	)
}
