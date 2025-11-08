package order

import (
	"fmt"
	"strings"

	"github.com/Adigezalov/gophermart/internal/utils"
)

// Service содержит бизнес-логику для заказов
type Service struct {
	repo Repository
}

// NewService создает новый экземпляр Service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// SubmitOrder загружает новый заказ пользователя
func (s *Service) SubmitOrder(userID int, orderNumber string) error {
	// Валидация номера заказа
	if err := s.validateOrderNumber(orderNumber); err != nil {
		return err
	}

	// Проверяем, существует ли уже заказ с таким номером
	existingOrder, err := s.repo.GetOrderByNumber(orderNumber)
	if err != nil && !strings.Contains(err.Error(), "заказ не найден") {
		return fmt.Errorf("ошибка при проверке существования заказа: %w", err)
	}

	// Если заказ уже существует
	if existingOrder != nil {
		if existingOrder.UserID == userID {
			// Заказ уже загружен этим пользователем
			return &OrderAlreadyExistsError{
				Message: "номер заказа уже был загружен этим пользователем",
				Code:    200,
			}
		} else {
			// Заказ загружен другим пользователем
			return &OrderConflictError{
				Message: "номер заказа уже был загружен другим пользователем",
				Code:    409,
			}
		}
	}

	// Создаем новый заказ
	newOrder := &Order{
		Number: orderNumber,
		UserID: userID,
		Status: StatusNew,
	}

	if err := s.repo.CreateOrder(newOrder); err != nil {
		// Проверяем на конфликт уникальности (race condition)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			// Пытаемся получить заказ еще раз
			existingOrder, getErr := s.repo.GetOrderByNumber(orderNumber)
			if getErr == nil && existingOrder != nil {
				if existingOrder.UserID == userID {
					return &OrderAlreadyExistsError{
						Message: "номер заказа уже был загружен этим пользователем",
						Code:    200,
					}
				} else {
					return &OrderConflictError{
						Message: "номер заказа уже был загружен другим пользователем",
						Code:    409,
					}
				}
			}
		}
		return fmt.Errorf("не удалось создать заказ: %w", err)
	}

	// Заказ успешно принят в обработку
	return &OrderAcceptedError{
		Message: "новый номер заказа принят в обработку",
		Code:    202,
	}
}

// validateOrderNumber валидирует номер заказа
func (s *Service) validateOrderNumber(orderNumber string) error {
	// Убираем пробелы
	orderNumber = strings.TrimSpace(orderNumber)

	// Проверяем, что номер не пустой
	if orderNumber == "" {
		return &OrderValidationError{
			Message: "номер заказа не может быть пустым",
			Code:    400,
		}
	}

	// Проверяем, что номер содержит только цифры
	for _, char := range orderNumber {
		if char < '0' || char > '9' {
			return &OrderValidationError{
				Message: "номер заказа должен содержать только цифры",
				Code:    422,
			}
		}
	}

	// Проверяем номер по алгоритму Луна
	if !utils.IsValidLuhn(orderNumber) {
		return &OrderValidationError{
			Message: "неверный формат номера заказа",
			Code:    422,
		}
	}

	return nil
}

// Кастомные ошибки для разных HTTP статусов

// OrderAlreadyExistsError - заказ уже загружен этим пользователем (200)
type OrderAlreadyExistsError struct {
	Message string
	Code    int
}

func (e *OrderAlreadyExistsError) Error() string {
	return e.Message
}

// OrderConflictError - заказ загружен другим пользователем (409)
type OrderConflictError struct {
	Message string
	Code    int
}

func (e *OrderConflictError) Error() string {
	return e.Message
}

// OrderAcceptedError - заказ принят в обработку (202)
type OrderAcceptedError struct {
	Message string
	Code    int
}

func (e *OrderAcceptedError) Error() string {
	return e.Message
}

// OrderValidationError - ошибка валидации (400/422)
type OrderValidationError struct {
	Message string
	Code    int
}

func (e *OrderValidationError) Error() string {
	return e.Message
}

// GetUserOrders получает все заказы пользователя
func (s *Service) GetUserOrders(userID int) ([]*Order, error) {
	orders, err := s.repo.GetOrdersByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить заказы пользователя: %w", err)
	}

	return orders, nil
}
