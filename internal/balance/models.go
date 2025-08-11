package balance

import (
	"time"
)

// Balance представляет баланс пользователя
type Balance struct {
	ID        int       `json:"-" db:"id"`
	UserID    int       `json:"-" db:"user_id"`
	Current   float64   `json:"current" db:"current"`
	Withdrawn float64   `json:"withdrawn" db:"withdrawn"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}