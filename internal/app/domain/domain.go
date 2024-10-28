package domain

import "time"

type User struct {
	UserID    string
	Login     string
	Password  uint32
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

type OrderStatus string

const (
	NewOrderStatus        OrderStatus = "NEW"
	ProcessingOrderStatus OrderStatus = "PROCESSING"
	ProcessedOrdersStatus OrderStatus = "PROCESSED"
	InvalidOrderStatus    OrderStatus = "INVALID"
)
