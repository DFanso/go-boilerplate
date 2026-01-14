package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"github.com/vidwadeseram/go-boilerplate/identity-api/gen/identity"
	db "github.com/vidwadeseram/go-boilerplate/identity-api/internal/db/sqlc"
	"github.com/vidwadeseram/go-boilerplate/identity-api/internal/security"
)

// Service implements the goa generated interface and orchestrates business logic.
type Service struct {
	log     *slog.Logger
	queries *db.Queries
	tokens  *security.TokenManager
}

// New creates a new Service instance.
func New(log *slog.Logger, queries *db.Queries, tokens *security.TokenManager) *Service {
	return &Service{log: log, queries: queries, tokens: tokens}
}

// Register creates a new user.
func (s *Service) Register(ctx context.Context, payload *identity.RegisterPayload) (*identity.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        payload.Email,
		PasswordHash: string(hashed),
		DisplayName:  payload.DisplayName,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("email already exists: %w", err)
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	result := mapUser(user)
	s.log.InfoContext(ctx, "registered user", "userID", result.ID)
	return result, nil
}

// Login validates credentials and issues a token.
func (s *Service) Login(ctx context.Context, payload *identity.Credentials) (*identity.TokenResult, error) {
	user, err := s.queries.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.WarnContext(ctx, "login failed: user not found", "email", payload.Email)
			return nil, &identity.UnauthorizedError{Message: "invalid credentials"}
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.Password)); err != nil {
		s.log.WarnContext(ctx, "login failed: password mismatch", "email", payload.Email)
		return nil, &identity.UnauthorizedError{Message: "invalid credentials"}
	}

	signed, ttl, err := s.tokens.Issue(user)
	if err != nil {
		return nil, fmt.Errorf("issue token: %w", err)
	}

	return &identity.TokenResult{AccessToken: signed, ExpiresIn: int(ttl.Seconds())}, nil
}

// ValidateToken verifies JWTs and exposes user identity.
func (s *Service) ValidateToken(ctx context.Context, payload *identity.ValidateTokenPayload) (*identity.ValidationResult, error) {
	claims, err := s.tokens.Validate(payload.Token)
	if err != nil {
		s.log.WarnContext(ctx, "token validation failed", "error", err)
		return &identity.ValidationResult{Valid: false, Reason: ptr(err.Error())}, nil
	}

	return &identity.ValidationResult{
		Valid:  true,
		UserID: ptr(claims.UserID),
		Email:  ptr(claims.Email),
	}, nil
}

func mapUser(u db.User) *identity.User {
	createdAt := time.Now().UTC()
	if u.CreatedAt.Valid {
		createdAt = u.CreatedAt.Time
	}

	id := ""
	if u.ID.Valid {
		id = u.ID.String()
	}

	return &identity.User{
		ID:          id,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		CreatedAt:   createdAt.UTC().Format(time.RFC3339),
	}
}

func ptr[T any](v T) *T {
	return &v
}
