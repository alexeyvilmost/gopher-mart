package storage

import (
	"context"
	"database/sql"

	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStorage struct {
	db *sql.DB
}

var initlist = map[string]string{
	"createUsers":            "CREATE TABLE IF NOT EXISTS users (login TEXT UNIQUE PRIMARY KEY, password TEXT, user_id TEXT UNIQUE, balance INTEGER);",
	"createOrders":           "CREATE TABLE IF NOT EXISTS orders (order_id TEXT UNIQUE PRIMARY KEY, user_id TEXT, accrual INTEGER, status TEXT, uploaded_at TIMESTAMP DEFAULT NOW());",
	"createWithdrawals":      "CREATE TABLE IF NOT EXISTS withdrawals (order_id TEXT UNIQUE PRIMARY KEY, user_id TEXT, sum INTEGER, processed_at TIMESTAMP);",
	"indexUsersUserId":       "CREATE INDEX IF NOT EXISTS users__user_id ON users (user_id);",
	"indexOrdersUserId":      "CREATE INDEX IF NOT EXISTS orders__user_id ON orders (user_id);",
	"indexWithdrawalsUserId": "CREATE INDEX IF NOT EXISTS withdrawals__user_id ON withdrawals (user_id);",
}

func NewDBStorage(conn string) (*DBStorage, error) {
	db, err := sql.Open("pgx", conn)
	if err != nil {
		return &DBStorage{}, fmt.Errorf("failed to create db from connection string: %w", err)
	}
	return &DBStorage{db}, nil
}

func (s *DBStorage) Init() error {
	for name, query := range initlist {
		_, err := s.db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to init during %s, err: %w", name, err)
		}
	}
	return nil
}

// row := s.db.QueryRowContext(ctx, "INSERT INTO urls VALUES ($1, $2, $3, FALSE) ON CONFLICT DO NOTHING RETURNING short_url;", shortURL, fullURL, userID)
// 	var str string
// 	err = row.Scan(&str)

func (s *DBStorage) AddUser(ctx context.Context, user User) (bool, error) {
	row := s.db.QueryRowContext(ctx, "INSERT INTO users VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING RETURNING user_id;", user.Login, user.Password, user.UserID, user.Balance)
	var str string
	if err := row.Scan(&str); err != nil {
		return false, err
	}
	return true, nil
}

func (s *DBStorage) GetUser(ctx context.Context, userID string) (User, error) {
	user := User{}
	row := s.db.QueryRowContext(ctx, "SELECT login, password, user_id, balance FROM users WHERE user_id = $1;", userID)
	if err := row.Scan(&user.Login, &user.Password, &user.UserID, &user.Balance); err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *DBStorage) GetUserID(ctx context.Context, login, password string) (string, error) {
	row := s.db.QueryRowContext(ctx, "SELECT user_id FROM users WHERE login = $1 AND password = $2;", login, password)
	var userID string
	if err := row.Scan(&userID); err != nil {
		return "", err
	}
	return userID, nil
}

func (s *DBStorage) CheckUser(ctx context.Context, login string) (exists bool, err error) {
	row := s.db.QueryRowContext(ctx, "SELECT user_id FROM users WHERE login = $1", login)
	var userID string
	if err := row.Scan(&userID); err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return len(userID) > 0, nil
}

func (s *DBStorage) UpdateUser(ctx context.Context, user User) (bool, error) {
	row := s.db.QueryRowContext(ctx, "UPDATE users SET balance = $1 WHERE user_id = $2 ON CONFLICT DO NOTHING;", user.Balance, user.UserID)
	if row.Err() != nil {
		return false, row.Err()
	}
	return true, nil
}
