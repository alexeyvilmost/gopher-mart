package storage

import (
	"context"
	"gophermart/internal/app/domain"
)

type Storage interface {
	Init() error

	AddUser(ctx context.Context, user domain.User) error
	GetUser(ctx context.Context, userID string) (domain.User, error)
	GetUserID(ctx context.Context, login string, password uint32) (string, error)
	CheckUser(ctx context.Context, login string) (exists bool, err error)
	UpdateUser(ctx context.Context, user domain.User) error

	AddOrder(ctx context.Context, order domain.Order) error
	GetOrders(ctx context.Context, userID string) ([]domain.Order, error)
	CheckOrder(ctx context.Context, userID, orderID string) (exists bool, err error)
	GetIncompleteOrders(ctx context.Context) ([]domain.Order, error)
	UpdateOrder(ctx context.Context, order domain.Order) error

	AddWithdrawal(ctx context.Context, wd domain.Withdrawal) error
	GetWithdrawals(ctx context.Context, userID string) ([]domain.Withdrawal, error)
}
