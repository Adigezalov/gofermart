package balance

import (
	"fmt"
	"log"
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

// GetUserBalance получает баланс пользователя
func (s *Service) GetUserBalance(userID int) (*Balance, error) {
	balance, err := s.repo.GetBalance(userID)
	if err != nil {
		// Если баланс не найден, создаем его
		if err.Error() == "баланс пользователя не найден" {
			if createErr := s.repo.CreateBalance(userID); createErr != nil {
				return nil, fmt.Errorf("не удалось создать баланс: %w", createErr)
			}
			// Повторно получаем созданный баланс
			balance, err = s.repo.GetBalance(userID)
			if err != nil {
				return nil, fmt.Errorf("не удалось получить созданный баланс: %w", err)
			}
		} else {
			return nil, fmt.Errorf("не удалось получить баланс пользователя: %w", err)
		}
	}

	return balance, nil
}

// AddPoints добавляет баллы к балансу пользователя
func (s *Service) AddPoints(userID int, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("сумма начисления должна быть положительной")
	}

	// Убеждаемся, что баланс существует
	_, err := s.GetUserBalance(userID)
	if err != nil {
		return fmt.Errorf("не удалось получить баланс для начисления: %w", err)
	}

	// Добавляем баллы (увеличиваем current, не изменяем withdrawn)
	err = s.repo.UpdateBalance(userID, amount, 0)
	if err != nil {
		log.Printf("Ошибка начисления %.2f баллов пользователю %d: %v", amount, userID, err)
		return fmt.Errorf("не удалось начислить баллы: %w", err)
	}

	log.Printf("Начислено %.2f баллов пользователю %d", amount, userID)
	return nil
}

// DeductPoints списывает баллы с баланса пользователя
func (s *Service) DeductPoints(userID int, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("сумма списания должна быть положительной")
	}

	// Получаем текущий баланс
	balance, err := s.GetUserBalance(userID)
	if err != nil {
		return fmt.Errorf("не удалось получить баланс для списания: %w", err)
	}

	// Проверяем достаточность средств
	if balance.Current < amount {
		log.Printf("Недостаточно средств для списания у пользователя %d: доступно %.2f, запрошено %.2f", 
			userID, balance.Current, amount)
		return &InsufficientFundsError{
			Available: balance.Current,
			Requested: amount,
		}
	}

	// Списываем баллы (уменьшаем current, увеличиваем withdrawn)
	err = s.repo.UpdateBalance(userID, -amount, amount)
	if err != nil {
		log.Printf("Ошибка списания %.2f баллов у пользователя %d: %v", amount, userID, err)
		return fmt.Errorf("не удалось списать баллы: %w", err)
	}

	log.Printf("Списано %.2f баллов у пользователя %d", amount, userID)
	return nil
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