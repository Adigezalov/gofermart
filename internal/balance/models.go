package balance

import (
	"time"
)

// Balance представляет баланс пользователя (в копейках для хранения/расчетов)
type Balance struct {
	ID             int       `json:"-" db:"id"`
	UserID         int       `json:"-" db:"user_id"`
	CurrentCents   int64     `json:"-" db:"current_cents"`
	WithdrawnCents int64     `json:"-" db:"withdrawn_cents"`
	UpdatedAt      time.Time `json:"-" db:"updated_at"`
}
