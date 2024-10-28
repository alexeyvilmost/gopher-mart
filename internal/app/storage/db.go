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
	"createUsers":       "CREATE TABLE IF NOT EXISTS users (login TEXT UNIQUE PRIMARY KEY, password TEXT, user_id TEXT UNIQUE, balance DECIMAL, withdrawn DECIMAL);",
	"createOrders":      "CREATE TABLE IF NOT EXISTS orders (order_id TEXT UNIQUE PRIMARY KEY, user_id TEXT, accrual INTEGER, status TEXT, uploaded_at TIMESTAMP DEFAULT NOW());",
	"createWithdrawals": "CREATE TABLE IF NOT EXISTS withdrawals (user_id TEXT UNIQUE PRIMARY KEY, order_id TEXT, sum INTEGER, processed_at TIMESTAMP DEFAULT NOW());",
	"indexUsersUserId":  "CREATE INDEX IF NOT EXISTS users__user_id ON users (user_id);",
	"indexOrdersUserId": "CREATE INDEX IF NOT EXISTS orders__user_id ON orders (user_id);",
}

var (
	ErrEmpty    = fmt.Errorf("empty db response")
	ErrConflict = fmt.Errorf("conflict")
)

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

func (s *DBStorage) AddUser(ctx context.Context, user User) error {
	row := s.db.QueryRowContext(ctx, "INSERT INTO users VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING RETURNING user_id;", user.Login, user.Password, user.UserID, user.Balance, user.Withdrawn)
	var str string
	if err := row.Scan(&str); err != nil {
		return err
	}
	return nil
}

func (s *DBStorage) GetUser(ctx context.Context, userID string) (User, error) {
	user := User{}
	row := s.db.QueryRowContext(ctx, "SELECT login, password, user_id, balance, withdrawn FROM users WHERE user_id = $1;", userID)
	if err := row.Scan(&user.Login, &user.Password, &user.UserID, &user.Balance, &user.Withdrawn); err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *DBStorage) GetUserID(ctx context.Context, login, password string) (string, error) {
	row := s.db.QueryRowContext(ctx, "SELECT user_id FROM users WHERE login = $1 AND password = $2;", login, password)
	var userID string
	if err := row.Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			return "", ErrEmpty
		}
		return "", err
	}
	return userID, nil
}

func (s *DBStorage) CheckUser(ctx context.Context, login string) (exists bool, err error) {
	row := s.db.QueryRowContext(ctx, "SELECT user_id FROM users WHERE login = $1;", login)
	var userID string
	if err := row.Scan(&userID); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil // TODO: remove unused bool's
}

func (s *DBStorage) UpdateUser(ctx context.Context, user User) error {
	row := s.db.QueryRowContext(ctx, "UPDATE users SET balance = $1 WHERE user_id = $2 ON CONFLICT DO NOTHING;", user.Balance, user.UserID)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

func (s *DBStorage) AddOrder(ctx context.Context, order Order) error {
	row := s.db.QueryRowContext(ctx, "INSERT INTO orders VALUES ($1, $2, $3, $4, NOW()) ON CONFLICT DO NOTHING RETURNING order_id;", order.OrderID, order.UserID, order.Accrual, string(order.Status))
	var str string
	if err := row.Scan(&str); err != nil {
		return err
	}
	return nil
}

func (s *DBStorage) UpdateOrder(ctx context.Context, order Order) error {
	row := s.db.QueryRowContext(ctx, "UPDATE orders SET status = $1, accrual = $2 WHERE order_id = $3 ON CONFLICT DO NOTHING;", order.Status, order.Accrual, order.OrderID)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

func (s *DBStorage) CheckOrder(ctx context.Context, userID, orderID string) (exists bool, err error) {
	row := s.db.QueryRowContext(ctx, "SELECT user_id FROM orders WHERE order_id = $1;", orderID)
	var orderUserID string
	if err := row.Scan(&orderUserID); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if orderUserID != userID {
		return true, ErrConflict
	}
	return true, nil // TODO: remove unused bool's
}

func (s *DBStorage) GetOrders(ctx context.Context, userID string) ([]Order, error) {
	result := []Order{}
	rows, err := s.db.QueryContext(ctx, "SELECT order_id, user_id, accrual, status, uploaded_at FROM orders WHERE user_id = $1;", userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		order := Order{}
		if err := rows.Scan(&order.OrderID, &order.UserID, &order.Accrual, &order.Status, &order.UploadedAt); err != nil {
			return nil, err
		}
		result = append(result, order)
	}

	return result, nil
}

func (s *DBStorage) GetIncompleteOrders(ctx context.Context) ([]Order, error) {
	result := []Order{}
	rows, err := s.db.QueryContext(ctx, "SELECT order_id, user_id, accrual, status, uploaded_at FROM orders WHERE status IN('NEW','PROCESSING');")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		order := Order{}
		if err := rows.Scan(&order.OrderID, &order.UserID, &order.Accrual, &order.Status, &order.UploadedAt); err != nil {
			return nil, err
		}
		result = append(result, order)
	}

	return result, nil
}

func (s *DBStorage) AddWithdrawal(ctx context.Context, wd Withdrawal) error {
	row := s.db.QueryRowContext(ctx, "INSERT INTO withdrawals VALUES ($1, $2, $3, NOW()) ON CONFLICT DO NOTHING RETURNING user_id;", wd.UserID, wd.OrderID, wd.Sum)
	var str string
	if err := row.Scan(&str); err != nil {
		return err
	}
	return nil
}

func (s *DBStorage) GetWithdrawals(ctx context.Context, userID string) ([]Withdrawal, error) {
	result := []Withdrawal{}
	rows, err := s.db.QueryContext(ctx, "SELECT order_id, user_id, sum, processed_at FROM withdrawals WHERE user_id = $1;", userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		wd := Withdrawal{}
		if err := rows.Scan(&wd.OrderID, &wd.UserID, &wd.Sum, &wd.ProcessedAt); err != nil {
			return nil, err
		}
		result = append(result, wd)
	}

	return result, nil
}
