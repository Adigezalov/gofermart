package withdrawal

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/Adigezalov/gophermart/internal/balance"
	"github.com/Adigezalov/gophermart/internal/utils"
)

// Service содержит бизнес-логику для операций списания
type Service struct {
	withdrawalRepo Repository
	balanceService *balance.Service
}

// NewService создает новый экземпляр Service
func NewService(withdrawalRepo Repository, balanceService *balance.Service) *Service {
	return &Service{
		withdrawalRepo: withdrawalRepo,
		balanceService: balanceService,
	}
}

// WithdrawPoints выполняет списание баллов (amount в рублях)
func (s *Service) WithdrawPoints(userID int, orderNumber string, amount float64) error {
	// Валидация номера заказа
	if err := s.validateOrderNumber(orderNumber); err != nil {
		return err
	}

	// Валидация суммы
	if amount <= 0 {
		return &ValidationError{
			Message: "сумма списания должна быть положительной",
			Code:    400,
		}
	}

	amountCents, err := rubToCents(amount)
	if err != nil {
		return &ValidationError{
			Message: "неверный формат суммы",
			Code:    400,
		}
	}

	// Попытка списания через balance service
	if err := s.balanceService.DeductPoints(userID, amount); err != nil {
		// Проверяем тип ошибки
		if insufficientErr, ok := err.(*balance.InsufficientFundsError); ok {
			return &InsufficientFundsError{
				Available: insufficientErr.Available,
				Requested: insufficientErr.Requested,
			}
		}
		return fmt.Errorf("не удалось списать баллы: %w", err)
	}

	// Создаем запись о списании (в копейках)
	withdrawal := &Withdrawal{
		UserID:      userID,
		OrderNumber: orderNumber,
		AmountCents: amountCents,
	}

	if err := s.withdrawalRepo.CreateWithdrawal(withdrawal); err != nil {
		// Если не удалось создать запись, пытаемся вернуть баллы (в рублях)
		if addErr := s.balanceService.AddPoints(userID, amount); addErr != nil {
			// Логируем критическую ошибку - баллы списаны, но запись не создана
			log.Printf("КРИТИЧЕСКАЯ ОШИБКА: не удалось вернуть баллы пользователю %d после неудачного списания: %v", userID, addErr)
		}
		log.Printf("Ошибка создания записи о списании для пользователя %d: %v", userID, err)
		return fmt.Errorf("не удалось создать запись о списании: %w", err)
	}

	log.Printf("Успешно списано %.2f баллов у пользователя %d за заказ %s", amount, userID, orderNumber)
	return nil
}

// GetUserWithdrawals получает историю списаний пользователя
// Возвращает модель с суммами в копейках, хендлер преобразует в рубли при ответе
func (s *Service) GetUserWithdrawals(userID int) ([]*Withdrawal, error) {
	withdrawals, err := s.withdrawalRepo.GetWithdrawalsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить историю списаний: %w", err)
	}
	return withdrawals, nil
}

// validateOrderNumber валидирует номер заказа
func (s *Service) validateOrderNumber(orderNumber string) error {
	// Убираем пробелы
	orderNumber = strings.TrimSpace(orderNumber)

	// Проверяем, что номер не пустой
	if orderNumber == "" {
		return &ValidationError{
			Message: "номер заказа не может быть пустым",
			Code:    400,
		}
	}

	// Проверяем, что номер содержит только цифры
	for _, char := range orderNumber {
		if char < '0' || char > '9' {
			return &ValidationError{
				Message: "номер заказа должен содержать только цифры",
				Code:    422,
			}
		}
	}

	// Проверяем номер по алгоритму Луна
	if !utils.IsValidLuhn(orderNumber) {
		return &ValidationError{
			Message: "неверный формат номера заказа",
			Code:    422,
		}
	}

	return nil
}

// Кастомные ошибки

// ValidationError ошибка валидации
type ValidationError struct {
	Message string
	Code    int
}

func (e *ValidationError) Error() string {
	return e.Message
}

// InsufficientFundsError ошибка недостаточности средств
type InsufficientFundsError struct {
	Available float64
	Requested float64
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("недостаточно средств: доступно %.2f, запрошено %.2f",
		e.Available, e.Requested)
}

// Локальная конвертация для сервиса списаний
func rubToCents(amount float64) (int64, error) {
	v := math.Round(amount * 100.0)
	if v > math.MaxInt64 || v < math.MinInt64 {
		return 0, fmt.Errorf("сумма выходит за пределы int64")
	}
	return int64(v), nil
}
