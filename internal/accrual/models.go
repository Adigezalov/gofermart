package accrual

import (
	"fmt"
)

// AccrualStatus статусы обработки начислений
type AccrualStatus string

const (
	StatusRegistered AccrualStatus = "REGISTERED" // заказ зарегистрирован, но начисление не рассчитано
	StatusInvalid    AccrualStatus = "INVALID"    // заказ не принят к расчёту, вознаграждение не будет начислено
	StatusProcessing AccrualStatus = "PROCESSING" // расчёт начисления в процессе
	StatusProcessed  AccrualStatus = "PROCESSED"  // расчёт начисления окончен
)

// AccrualResponse ответ от системы начислений
type AccrualResponse struct {
	Order   string        `json:"order"`
	Status  AccrualStatus `json:"status"`
	Accrual *float64      `json:"accrual,omitempty"`
}

// IsValid проверяет валидность статуса
func (s AccrualStatus) IsValid() bool {
	switch s {
	case StatusRegistered, StatusInvalid, StatusProcessing, StatusProcessed:
		return true
	default:
		return false
	}
}

// IsFinal проверяет, является ли статус финальным
func (s AccrualStatus) IsFinal() bool {
	return s == StatusInvalid || s == StatusProcessed
}

// Validate валидирует ответ от системы начислений
func (r *AccrualResponse) Validate() error {
	if r.Order == "" {
		return fmt.Errorf("номер заказа не может быть пустым")
	}

	if !r.Status.IsValid() {
		return fmt.Errorf("неизвестный статус: %s", r.Status)
	}

	// Для статуса PROCESSED начисление может отсутствовать (если нет начислений за заказ)
	// Это нормальная ситуация согласно спецификации

	// Для других статусов начисление должно отсутствовать или быть нулевым
	if r.Status != StatusProcessed && r.Accrual != nil && *r.Accrual > 0 {
		return fmt.Errorf("начисление должно быть указано только для статуса PROCESSED")
	}

	return nil
}
