package withdrawal

import (
	"time"
)

// Withdrawal представляет операцию списания баллов
type Withdrawal struct {
	ID          int       `json:"-" db:"id"`
	UserID      int       `json:"-" db:"user_id"`
	OrderNumber string    `json:"order" db:"order_number"`
	Amount      float64   `json:"sum" db:"amount"`
	ProcessedAt time.Time `json:"processed_at" db:"processed_at"`
}

// WithdrawRequest представляет запрос на списание баллов
type WithdrawRequest struct {
	OrderNumber string  `json:"order"`
	Amount      float64 `json:"sum"`
}