package withdrawal

import (
	"time"
)

// Withdrawal представляет операцию списания баллов (хранение в копейках)
type Withdrawal struct {
	ID          int       `json:"-" db:"id"`
	UserID      int       `json:"-" db:"user_id"`
	OrderNumber string    `json:"order" db:"order_number"`
	AmountCents int64     `json:"-" db:"amount_cents"`
	ProcessedAt time.Time `json:"processed_at" db:"processed_at"`
}

// WithdrawRequest представляет запрос на списание баллов (API в рублях)
type WithdrawRequest struct {
	OrderNumber string  `json:"order"`
	Amount      float64 `json:"sum"`
}
