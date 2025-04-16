package store

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"time"
)

type PostStore struct {
	db *sqlx.DB
}

func NewPostStore(db *sql.DB) *PostStore {
	return &PostStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type Posts struct {
	UserId    int       `db:"user_id"`
	Id        int       `db:"id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (s *PostStore) CreatePost(ctx context.Context, title, content string) (*Posts, error) {
	dml := `INSERT INTO posts (user_id, id, title, content) VALUES ($1, $2, $3, $4) RETURNING *`
	var post Posts
	userId := ctx.Value("user").(*User).Id

	// TODO check why id is required in postgres
	if err := s.db.GetContext(ctx, &post, dml, userId, 2, title, content); err != nil {
		return nil, fmt.Errorf("failed to insert post: %w", err)
	}
	return &post, nil
}
