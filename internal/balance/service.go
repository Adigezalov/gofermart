package balance

import (
	"fmt"
	"log"
	"math"
	"strings"
)

// Service содержит бизнес-логику для балансов
type Service struct {
	repo Repository
}

// NewService создает новый экземпляр Service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// EnsureBalance гарантирует наличие начального баланса для пользователя.
// Если запись уже существует (duplicate key), не считается ошибкой.
func (s *Service) EnsureBalance(userID int) error {
	if err := s.repo.CreateBalance(userID); err != nil {
		// Игнорируем конфликт уникальности (баланс уже есть)
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(strings.ToLower(err.Error()), "unique") ||
			strings.Contains(strings.ToLower(err.Error()), "уник") {
			return nil
		}
		return fmt.Errorf("не удалось создать баланс: %w", err)
	}
	return nil
}

// GetUserBalance получает баланс пользователя
func (s *Service) GetUserBalance(userID int) (*Balance, error) {
	bal, err := s.repo.GetBalance(userID)
	if err != nil {
		// Если баланс не найден, создаем его
		if err.Error() == "баланс пользователя не найден" {
			if createErr := s.repo.CreateBalance(userID); createErr != nil {
				return nil, fmt.Errorf("не удалось создать баланс: %w", createErr)
			}
			// Повторно получаем созданный баланс
			bal, err = s.repo.GetBalance(userID)
			if err != nil {
				return nil, fmt.Errorf("не удалось получить созданный баланс: %w", err)
			}
		} else {
			return nil, fmt.Errorf("не удалось получить баланс пользователя: %w", err)
		}
	}
	return bal, nil
}

// AddPoints добавляет баллы к балансу пользователя (amount в рублях)
func (s *Service) AddPoints(userID int, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("сумма начисления должна быть положительной")
	}

	// Убеждаемся, что баланс существует
	_, err := s.GetUserBalance(userID)
	if err != nil {
		return fmt.Errorf("не удалось получить баланс для начисления: %w", err)
	}

	amountCents, err := rubToCents(amount)
	if err != nil {
		return fmt.Errorf("неверная сумма начисления: %w", err)
	}

	// Добавляем баллы (увеличиваем current_cents, не изменяем withdrawn_cents)
	err = s.repo.UpdateBalance(userID, amountCents, 0)
	if err != nil {
		log.Printf("Ошибка начисления %.2f баллов пользователю %d: %v", amount, userID, err)
		return fmt.Errorf("не удалось начислить баллы: %w", err)
	}

	log.Printf("Начислено %.2f баллов пользователю %d", amount, userID)
	return nil
}

// DeductPoints списывает баллы с баланса пользователя (amount в рублях)
func (s *Service) DeductPoints(userID int, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("сумма списания должна быть положительной")
	}

	// Получаем текущий баланс
	bal, err := s.GetUserBalance(userID)
	if err != nil {
		return fmt.Errorf("не удалось получить баланс для списания: %w", err)
	}

	amountCents, err := rubToCents(amount)
	if err != nil {
		return fmt.Errorf("неверная сумма списания: %w", err)
	}

	// Проверяем достаточность средств (в копейках)
	if bal.CurrentCents < amountCents {
		availableRub := centsToRub(bal.CurrentCents)
		log.Printf("Недостаточно средств для списания у пользователя %d: доступно %.2f, запрошено %.2f",
			userID, availableRub, amount)
		return &InsufficientFundsError{
			Available: availableRub,
			Requested: amount,
		}
	}

	// Списываем баллы (уменьшаем current_cents, увеличиваем withdrawn_cents)
	err = s.repo.UpdateBalance(userID, -amountCents, amountCents)
	if err != nil {
		log.Printf("Ошибка списания %.2f баллов у пользователя %d: %v", amount, userID, err)
		return fmt.Errorf("не удалось списать баллы: %w", err)
	}

	log.Printf("Списано %.2f баллов у пользователя %d", amount, userID)
	return nil
}

// Вспомогательные функции конвертации

func rubToCents(amount float64) (int64, error) {
	// Округление до ближайшей копейки
	v := math.Round(amount * 100.0)
	if v > float64(math.MaxInt64) || v < float64(math.MinInt64) {
		return 0, fmt.Errorf("сумма выходит за пределы int64")
	}
	return int64(v), nil
}

func centsToRub(cents int64) float64 {
	return float64(cents) / 100.0
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
