package store

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UsersStore struct {
	db *sqlx.DB
}

func NewUsersStore(db *sql.DB) *UsersStore {
	return &UsersStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type User struct {
	Id                   int       `db:"id"`
	Email                string    `db:"email"`
	Username             string    `db:"username"`
	HashedPasswordBase64 string    `db:"hashed_password"`
	CreatedAt            time.Time `db:"created_at"`
}

func (u *User) CheckHashedPassword(password string) error {
	hashedPassword, err := base64.StdEncoding.DecodeString(u.HashedPasswordBase64)
	if err != nil {
		return fmt.Errorf("failed to decode hashed password: %w", err)
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		return fmt.Errorf("password invalid")
	}
	return nil
}

func (s *UsersStore) CreateUser(ctx context.Context, username, email, password string) (*User, error) {
	dml := `INSERT INTO users (username, email, hashed_password) VALUES ($1, $2, $3) RETURNING *`
	var user User

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}

	hashedPasswordBase64 := base64.StdEncoding.EncodeToString(hashedPassword)

	if err := s.db.GetContext(ctx, &user, dml, username, email, hashedPasswordBase64); err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}
	return &user, nil
}

func (s *UsersStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT * FROM users WHERE email = $1`
	var user User
	if err := s.db.GetContext(ctx, &user, query, email); err != nil {
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}
	return &user, nil
}

func (s *UsersStore) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `SELECT * FROM users WHERE username = $1`
	var user User
	if err := s.db.GetContext(ctx, &user, query, username); err != nil {
		return nil, fmt.Errorf("failed to query user by username: %w", err)
	}
	return &user, nil
}

func (s *UsersStore) GetUserById(ctx context.Context, id int) (*User, error) {
	query := `SELECT * FROM users WHERE id = $1`
	var user User
	if err := s.db.GetContext(ctx, &user, query, id); err != nil {
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}
	return &user, nil
}
