package domain

import (
	"encoding/json"
	"time"
)

func (o *Order) MarshalJSON() ([]byte, error) {
	var accrual *float64
	if o.Accrual == 0 {
		accrual = nil
	} else {
		accrual = &o.Accrual
	}
	return json.Marshal(&struct {
		OrderID    string      `json:"order_id"`
		Accrual    *float64    `json:"accrual,omitempty"`
		Status     OrderStatus `json:"status"`
		UploadedAt string      `json:"uploaded_at"`
	}{
		OrderID:    o.OrderID,
		Accrual:    accrual,
		Status:     o.Status,
		UploadedAt: o.UploadedAt.Format(time.RFC3339),
	})
}

func (w *Withdrawal) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		OrderID     string  `json:"order_id"`
		Sum         float64 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}{
		OrderID:     w.OrderID,
		Sum:         w.Sum,
		ProcessedAt: w.ProcessedAt.Format(time.RFC3339),
	})
}
