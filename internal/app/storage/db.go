package storage

import (
	"context"
	"database/sql"
	"fmt"
)

type DBStorage struct {
	db *sql.DB
}

var initlist = map[string]string{
	"createUsers":            "CREATE TABLE IF NOT EXISTS users (login TEXT UNIQUE PRIMARY KEY, password TEXT, balance INTEGER, user_id TEXT UNIQUE);",
	"createOrders":           "CREATE TABLE IF NOT EXISTS orders (order_id TEXT UNIQUE PRIMARY KEY, userID TEXT, accrual INTEGER, status TEXT, uploaded_at DATETIME DEFAULT NOW());",
	"createWithdrawals":      "CREATE TABLE IF NOT EXISTS withdrawal (order_id TEXT UNIQUE PRIMARY KEY, userID TEXT, sum INTEGER, processed_at DATETIME);",
	"indexUsersUserId":       "CREATE INDEX IF NOT EXISTS users__user_id ON users (user_id);",
	"indexOrdersUserId":      "CREATE INDEX IF NOT EXISTS orders__user_id ON orders (user_id);",
	"indexWithdrawalsUserId": "CREATE INDEX IF NOT EXISTS withdrawals__user_id ON withdrawals (user_id)",
}

func NewDBStorage(conn string) (*DBStorage, error) {
	db, err := sql.Open("pgx", conn)
	if err != nil {
		return &DBStorage{}, fmt.Errorf("failed to create db from connection string: %w", err)
	}
	return &DBStorage{db}, nil
}

func (s *DBStorage) Init(ctx context.Context) error {
	for name, query := range initlist {
		_, err := s.db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to init during %s, err: %w", name, err)
		}
	}
	return nil
}
