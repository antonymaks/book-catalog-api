package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const claimsContextKey contextKey = "auth_claims"

func (s *Service) Middleware(
	next http.Handler,
) http.Handler {

	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {

			header :=
				r.Header.Get("Authorization")

			if header == "" {
				http.Error(
					w,
					`{"error":"authorization required"}`,
					http.StatusUnauthorized,
				)
				return
			}

			const prefix = "Bearer "

			if !strings.HasPrefix(
				header,
				prefix,
			) {
				http.Error(
					w,
					`{"error":"invalid authorization header"}`,
					http.StatusUnauthorized,
				)
				return
			}

			tokenString :=
				strings.TrimSpace(
					strings.TrimPrefix(
						header,
						prefix,
					),
				)

			if tokenString == "" {
				http.Error(
					w,
					`{"error":"token is required"}`,
					http.StatusUnauthorized,
				)
				return
			}

			claims, err :=
				s.ParseToken(tokenString)

			if err != nil {
				http.Error(
					w,
					`{"error":"invalid or expired token"}`,
					http.StatusUnauthorized,
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				claimsContextKey,
				claims,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

func ClaimsFromContext(
	ctx context.Context,
) (*Claims, bool) {

	claims, ok :=
		ctx.Value(
			claimsContextKey,
		).(*Claims)

	return claims, ok
}

func UserIDFromContext(
	ctx context.Context,
) (int64, bool) {

	claims, ok :=
		ClaimsFromContext(ctx)

	if !ok {
		return 0, false
	}

	return claims.UserID, true
}

func RoleFromContext(
	ctx context.Context,
) (string, bool) {

	claims, ok :=
		ClaimsFromContext(ctx)

	if !ok {
		return "", false
	}

	return claims.Role, true
}

func (s *Service) OptionalMiddleware(
	next http.Handler,
) http.Handler {

	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			header := r.Header.Get("Authorization")

			if header == "" {
				next.ServeHTTP(w, r)
				return
			}

			const prefix = "Bearer "

			if !strings.HasPrefix(header, prefix) {
				next.ServeHTTP(w, r)
				return
			}

			tokenString := strings.TrimSpace(
				strings.TrimPrefix(
					header,
					prefix,
				),
			)

			if tokenString == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := s.ParseToken(tokenString)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				claimsContextKey,
				claims,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

func (s *Service) RequireRole(
	role string,
	next http.Handler,
) http.Handler {

	return s.Middleware(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {

				currentRole, ok :=
					RoleFromContext(r.Context())

				if !ok || currentRole != role {
					http.Error(
						w,
						"forbidden",
						http.StatusForbidden,
					)

					return
				}

				next.ServeHTTP(w, r)
			},
		),
	)
}
