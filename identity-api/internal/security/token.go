package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	db "github.com/vidwadeseram/go-boilerplate/identity-api/internal/db/sqlc"
)

// TokenManager issues and validates JWTs.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

// Claims represent the validated JWT claims used by other services.
type Claims struct {
	UserID string
	Email  string
}

// NewTokenManager builds a new TokenManager instance.
func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	if ttl <= 0 {
		ttl = time.Hour
	}
	return &TokenManager{secret: []byte(secret), ttl: ttl}
}

// Issue creates a signed JWT for the given user.
func (m *TokenManager) Issue(user db.User) (string, time.Duration, error) {
	userID := ""
	if user.ID.Valid {
		userID = user.ID.String()
	}
	if userID == "" {
		return "", 0, fmt.Errorf("user id is empty")
	}

	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(m.ttl)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   claims.Subject,
		"iat":   claims.IssuedAt.Unix(),
		"exp":   claims.ExpiresAt.Unix(),
		"email": user.Email,
	})

	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", 0, fmt.Errorf("sign token: %w", err)
	}

	return signed, m.ttl, nil
}

// Validate parses and validates a JWT string.
func (m *TokenManager) Validate(tokenString string) (*Claims, error) {
	parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, fmt.Errorf("token invalid")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("unexpected claims type")
	}

	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	if sub == "" {
		return nil, fmt.Errorf("token missing subject")
	}

	return &Claims{UserID: sub, Email: email}, nil
}
