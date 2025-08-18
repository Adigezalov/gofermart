package tokens

import (
	"time"
)

// RefreshToken представляет refresh токен в БД
type RefreshToken struct {
	ID        int       `json:"id" db:"id"`
	Token     string    `json:"token" db:"token"`
	UserID    int       `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TokenPair представляет пару токенов для ответа клиенту
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Claims представляет данные в JWT токене
type Claims struct {
	UserID int    `json:"user_id"`
	Login  string `json:"login"`
	Type   string `json:"type"` // "access" или "refresh"
}
