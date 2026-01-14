package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/vidwadeseram/go-boilerplate/dummy-api/gen/dummy"
	"github.com/vidwadeseram/go-boilerplate/dummy-api/internal/auth"
	db "github.com/vidwadeseram/go-boilerplate/dummy-api/internal/db/sqlc"
)

// IdentityValidator exposes token validation behavior.
type IdentityValidator interface {
	Validate(ctx context.Context, token string) (*auth.Claims, error)
}

// Service wires transports to business logic and data stores.
type Service struct {
	log       *slog.Logger
	queries   *db.Queries
	validator IdentityValidator
}

// New creates a new Service instance.
func New(log *slog.Logger, queries *db.Queries, validator IdentityValidator) *Service {
	return &Service{log: log, queries: queries, validator: validator}
}

// CreateItem inserts a new record for the authenticated user.
func (s *Service) CreateItem(ctx context.Context, payload *dummy.CreateItemPayload) (*dummy.Item, error) {
	claims, err := s.authorize(ctx, payload.Token)
	if err != nil {
		return nil, err
	}

	ownerID, err := toUUID(claims.UserID)
	if err != nil {
		return nil, &dummy.DummyUnauthorizedError{Message: "invalid subject claim"}
	}

	item, err := s.queries.CreateItem(ctx, db.CreateItemParams{
		OwnerID:     ownerID,
		Name:        payload.Name,
		Description: payload.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("create item: %w", err)
	}

	result := mapItem(item)
	s.log.InfoContext(ctx, "created item", "itemID", result.ID, "ownerID", result.OwnerID)
	return result, nil
}

// ListItems returns all entries for the authenticated user.
func (s *Service) ListItems(ctx context.Context, payload *dummy.ListItemsPayload) (*dummy.ItemsCollection, error) {
	claims, err := s.authorize(ctx, payload.Token)
	if err != nil {
		return nil, err
	}

	ownerID, err := toUUID(claims.UserID)
	if err != nil {
		return nil, &dummy.DummyUnauthorizedError{Message: "invalid subject claim"}
	}

	rows, err := s.queries.ListItems(ctx, db.ListItemsParams{OwnerID: ownerID, Limit: 50, Offset: 0})
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}

	items := make([]*dummy.Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapItem(row))
	}

	return &dummy.ItemsCollection{Items: items}, nil
}

// GetItem fetches a single entry if owned by the caller.
func (s *Service) GetItem(ctx context.Context, payload *dummy.ItemIDPayload) (*dummy.Item, error) {
	claims, err := s.authorize(ctx, payload.Token)
	if err != nil {
		return nil, err
	}

	ownerID, err := toUUID(claims.UserID)
	if err != nil {
		return nil, &dummy.DummyUnauthorizedError{Message: "invalid subject claim"}
	}

	itemID, err := toUUID(payload.ID)
	if err != nil {
		return nil, &dummy.DummyNotFoundError{Message: "invalid item id"}
	}

	item, err := s.queries.GetItem(ctx, db.GetItemParams{ID: itemID, OwnerID: ownerID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &dummy.DummyNotFoundError{Message: "item not found"}
		}
		return nil, fmt.Errorf("get item: %w", err)
	}

	return mapItem(item), nil
}

// DeleteItem removes an entry owned by the authenticated user.
func (s *Service) DeleteItem(ctx context.Context, payload *dummy.ItemIDPayload) error {
	claims, err := s.authorize(ctx, payload.Token)
	if err != nil {
		return err
	}

	ownerID, err := toUUID(claims.UserID)
	if err != nil {
		return &dummy.DummyUnauthorizedError{Message: "invalid subject claim"}
	}

	itemID, err := toUUID(payload.ID)
	if err != nil {
		return &dummy.DummyNotFoundError{Message: "invalid item id"}
	}

	if err := s.queries.DeleteItem(ctx, db.DeleteItemParams{ID: itemID, OwnerID: ownerID}); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	s.log.InfoContext(ctx, "deleted item", "itemID", payload.ID, "ownerID", claims.UserID)
	return nil
}

func (s *Service) authorize(ctx context.Context, raw string) (*auth.Claims, error) {
	token := extractToken(raw)
	if token == "" {
		return nil, &dummy.DummyUnauthorizedError{Message: "missing Authorization header"}
	}

	claims, err := s.validator.Validate(ctx, token)
	if err != nil {
		return nil, &dummy.DummyUnauthorizedError{Message: err.Error()}
	}

	return claims, nil
}

func extractToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parts := strings.SplitN(value, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return value
}

func toUUID(id string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return pgtype.UUID{}, err
	}
	var result pgtype.UUID
	copy(result.Bytes[:], parsed[:])
	result.Valid = true
	return result, nil
}

func mapItem(item db.Item) *dummy.Item {
	createdAt := time.Now().UTC()
	if item.CreatedAt.Valid {
		createdAt = item.CreatedAt.Time
	}

	owner := ""
	if item.OwnerID.Valid {
		owner = item.OwnerID.String()
	}

	id := ""
	if item.ID.Valid {
		id = item.ID.String()
	}

	return &dummy.Item{
		ID:          id,
		Name:        item.Name,
		Description: item.Description,
		OwnerID:     owner,
		CreatedAt:   createdAt.UTC().Format(time.RFC3339),
	}
}
