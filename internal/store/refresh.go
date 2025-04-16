package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log/slog"
	"time"
)

type RefreshTokenStore struct {
	db *sqlx.DB
}

func NewRefreshTokenStore(db *sql.DB) *RefreshTokenStore {
	return &RefreshTokenStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type RefreshToken struct {
	UserId      int       `db:"user_id"`
	HashedToken string    `db:"hashed_token"`
	CreatedAt   time.Time `db:"created_at"`
	ExpiresAt   time.Time `db:"expires_at"`
}

func (s *RefreshTokenStore) getBase64HashFromToken(token *jwt.Token) (string, error) {
	hashedToken := sha256.New()
	hashedToken.Write([]byte(token.Raw))
	hashedBytes := hashedToken.Sum(nil)
	base64HashedToken := base64.StdEncoding.EncodeToString(hashedBytes)
	return base64HashedToken, nil
}

func (s *RefreshTokenStore) CreateToken(ctx context.Context, userId int, token *jwt.Token) (*RefreshToken, error) {
	dml := `INSERT INTO refresh_tokens (user_id, hashed_token, expires_at) VALUES ($1, $2, $3)`
	base64HashedToken, err := s.getBase64HashFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 hashed token: %w", err)
	}

	expiresAt, err := token.Claims.GetExpirationTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get expiration time: %w", err)
	}

	var refreshToken RefreshToken
	if err := s.db.GetContext(ctx, &refreshToken, dml, userId, base64HashedToken, expiresAt.Time); err != nil {
		return nil, fmt.Errorf("failed to insert refresh token: %w", err)
	}

	return &refreshToken, nil
}

func (s *RefreshTokenStore) DeleteToken(ctx context.Context, userId int) (sql.Result, error) {
	dml := `DELETE FROM refresh_tokens WHERE user_id = $1`
	result, err := s.db.ExecContext(ctx, dml, userId)
	if err != nil {
		slog.Error("failed to delete refresh token", "err", err)
		return nil, fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return result, nil
}

func (s *RefreshTokenStore) GetToken(ctx context.Context, token *jwt.Token) (*RefreshToken, error) {
	query := `SELECT * FROM refresh_tokens WHERE user_id = $1 AND hashed_token = $2`
	base64HashedToken, err := s.getBase64HashFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 hashed token: %w", err)
	}
	var refreshToken RefreshToken
	if err := s.db.GetContext(ctx, &refreshToken, query, base64HashedToken); err != nil {
		return nil, fmt.Errorf("failed to fetch refresh token: %w", err)
	}
	return &refreshToken, nil
}
