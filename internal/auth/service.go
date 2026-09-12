package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"book-catalog-api/internal/apperror"
	"book-catalog-api/internal/domain"
	"book-catalog-api/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	users     repository.UserRepository
	jwtSecret []byte
}

type RegisterInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResult struct {
	Token string      `json:"token"`
	User  domain.User `json:"user"`
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`

	jwt.RegisteredClaims
}

func NewService(
	users repository.UserRepository,
	jwtSecret string,
) *Service {
	return &Service{
		users:     users,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *Service) Register(
	ctx context.Context,
	input RegisterInput,
) (*AuthResult, error) {

	username := strings.TrimSpace(input.Username)
	password := input.Password

	if len(username) < 3 {
		return nil, fmt.Errorf("username must contain at least 3 characters")
	}

	if len(password) < 6 {
		return nil, fmt.Errorf("password must contain at least 6 characters")
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to hash password: %w",
			err,
		)
	}

	user, err := s.users.Create(
		ctx,
		domain.CreateUserRequest{
			Username:     username,
			PasswordHash: string(hash),
			Role:         "user",
		},
	)

	if err != nil {
		return nil, err
	}

	token, err := s.generateToken(*user)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		Token: token,
		User:  *user,
	}, nil
}

func (s *Service) Login(
	ctx context.Context,
	input LoginInput,
) (*AuthResult, error) {

	username := strings.TrimSpace(input.Username)

	user, err := s.users.GetByUsername(
		ctx,
		username,
	)

	if errors.Is(err, apperror.ErrNotFound) {
		return nil, apperror.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(input.Password),
	)

	if err != nil {
		return nil, apperror.ErrNotFound
	}

	publicUser := domain.User{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	}

	token, err := s.generateToken(publicUser)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		Token: token,
		User:  publicUser,
	}, nil
}

func (s *Service) ParseToken(
	tokenString string,
) (*Claims, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(
					"unexpected signing method",
				)
			}

			return s.jwtSecret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)

	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (s *Service) generateToken(
	user domain.User,
) (string, error) {

	now := time.Now()

	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,

		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(now),

			ExpiresAt: jwt.NewNumericDate(
				now.Add(24 * time.Hour),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	tokenString, err := token.SignedString(
		s.jwtSecret,
	)

	if err != nil {
		return "", fmt.Errorf(
			"failed to create token: %w",
			err,
		)
	}

	return tokenString, nil
}
