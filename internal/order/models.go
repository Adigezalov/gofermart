package order

import (
	"time"
)

// OrderStatus представляет статус обработки заказа
type OrderStatus string

const (
	StatusNew        OrderStatus = "NEW"        // заказ загружен в систему, но не попал в обработку
	StatusProcessing OrderStatus = "PROCESSING" // вознаграждение за заказ рассчитывается
	StatusInvalid    OrderStatus = "INVALID"    // система расчёта вознаграждений отказала в расчёте
	StatusProcessed  OrderStatus = "PROCESSED"  // данные по заказу проверены и информация о расчёте успешно получена
)

// Order представляет заказ пользователя
type Order struct {
	ID         int         `json:"-" db:"id"`
	Number     string      `json:"number" db:"number"`
	UserID     int         `json:"-" db:"user_id"`
	Status     OrderStatus `json:"status" db:"status"`
	Accrual    *float64    `json:"accrual,omitempty" db:"accrual"`
	UploadedAt time.Time   `json:"uploaded_at" db:"uploaded_at"`
	UpdatedAt  time.Time   `json:"-" db:"updated_at"`
}

// SubmitOrderRequest представляет запрос на загрузку заказа
// Номер заказа передается как text/plain в теле запроса
type SubmitOrderRequest struct {
	Number string
}
