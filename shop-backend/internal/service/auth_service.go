package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/shophub/shop/internal/middleware"
	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
)

var (
	ErrEmailTaken     = errors.New("email already in use")
	ErrInvalidCreds   = errors.New("invalid email or password")
)

type AuthService struct {
	users     repository.UserRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewAuthService(users repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  24 * time.Hour,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         model.RoleUser,
	}

	if err := s.users.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *model.User, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		return "", nil, ErrInvalidCreds
	}
	if err != nil {
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCreds
	}

	token, err := s.issueToken(user)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (s *AuthService) Me(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.users.FindByID(ctx, id)
}

// SeedAdmin creates an admin user if one does not already exist.
func (s *AuthService) SeedAdmin(ctx context.Context, email, password string) error {
	_, err := s.users.FindByEmail(ctx, email)
	if err == nil {
		return nil // already exists
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.Create(ctx, &model.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         model.RoleAdmin,
	})
}

func (s *AuthService) issueToken(user *model.User) (string, error) {
	claims := &middleware.Claims{
		UserID: user.ID.String(),
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}
