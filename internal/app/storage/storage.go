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
	UserID    string
	Login     string
	Password  string
	Balance   float64
	Withdrawn float64
}

type Order struct {
	OrderID    string      `json:"order_id"`
	UserID     string      `json:"-"`
	Accrual    float64     `json:"accrual"`
	Status     OrderStatus `json:"status"`
	UploadedAt time.Time   `json:"uploaded_at"`
}

type Withdrawal struct {
	UserID      string    `json:"-"`
	OrderID     string    `json:"order_id"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

// TODO: Move all structs to domain.

type Storage interface {
	Init() error

	AddUser(ctx context.Context, user User) error
	GetUser(ctx context.Context, userID string) (User, error)
	GetUserID(ctx context.Context, login, password string) (string, error)
	CheckUser(ctx context.Context, login string) (exists bool, err error)
	UpdateUser(ctx context.Context, user User) error

	AddOrder(ctx context.Context, order Order) error
	GetOrders(ctx context.Context, userID string) ([]Order, error)
	CheckOrder(ctx context.Context, userID, orderID string) (exists bool, err error)
	// UpdateOrder(ctx context.Context, order Order) (bool, error)

	AddWithdrawal(ctx context.Context, wd Withdrawal) error
	GetWithdrawals(ctx context.Context, userID string) ([]Withdrawal, error)
	// UpdateWithdrawal(ctx context.Context, wd Withdrawal) (bool, error)
}
