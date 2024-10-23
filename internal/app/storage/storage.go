package storage

import (
	"context"
	"time"
)

type OrderStatus string

const (
	NewOrderStatus        OrderStatus = "NEW"
	ProcessingOrderStatus OrderStatus = "PROCESSING"
	ProcessedOrdersStatus OrderStatus = "PROCESSED"
	InvalidOrderStatus    OrderStatus = "INVALID"
)

type User struct {
	userID   string
	login    string
	password string
	balance  int64
}

type Order struct {
	orderID    string
	userID     string
	accrual    int64
	status     OrderStatus
	uploadedAt time.Time
}

type Withdrawal struct {
	userID      string
	orderID     string
	sum         int
	processedAt time.Time
}

type Storage interface {
	Init(ctx context.Context) error

	AddUser(ctx context.Context, user User) (bool, error)
	GetUser(ctx context.Context, userID string) (User, error)
	GetUserID(ctx context.Context, login, password string) (string, error)
	UpdateUser(ctx context.Context, user User) (bool, error)

	AddOrder(ctx context.Context, order Order) (bool, error)
	GetOrders(ctx context.Context, userID string) ([]Order, error)
	UpdateOrder(ctx context.Context, order Order) (bool, error)

	AddWithdrawal(ctx context.Context, wd Withdrawal) (bool, error)
	GetWithdrawals(ctx context.Context, userID string) ([]Withdrawal, error)
	UpdateWithdrawal(ctx context.Context, wd Withdrawal) (bool, error)
}
