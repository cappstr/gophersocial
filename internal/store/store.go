package store

import "database/sql"

type Store struct {
	User    *UsersStore
	Refresh *RefreshTokenStore
	Posts   *PostStore
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		User:    NewUsersStore(db),
		Refresh: NewRefreshTokenStore(db),
		Posts:   NewPostStore(db),
	}
}
