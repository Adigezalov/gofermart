package balance

import (
	"database/sql"
	"fmt"
)

// Repository интерфейс для работы с балансами в БД
type Repository interface {
	GetBalance(userID int) (*Balance, error)
	UpdateBalance(userID int, currentDeltaCents, withdrawnDeltaCents int64) error
	CreateBalance(userID int) error
}

// DatabaseRepository реализация Repository для PostgreSQL
type DatabaseRepository struct {
	db *sql.DB
}

// NewDatabaseRepository создает новый экземпляр DatabaseRepository
func NewDatabaseRepository(db *sql.DB) *DatabaseRepository {
	return &DatabaseRepository{db: db}
}

// GetBalance получает баланс пользователя (из *_cents)
func (r *DatabaseRepository) GetBalance(userID int) (*Balance, error) {
	balance := &Balance{}
	query := `
		SELECT id, user_id, current_cents, withdrawn_cents, updated_at 
		FROM user_balances 
		WHERE user_id = $1`

	err := r.db.QueryRow(query, userID).Scan(
		&balance.ID, &balance.UserID, &balance.CurrentCents,
		&balance.WithdrawnCents, &balance.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("баланс пользователя не найден")
		}
		return nil, fmt.Errorf("не удалось получить баланс: %w", err)
	}

	return balance, nil
}

// UpdateBalance обновляет баланс пользователя атомарно в копейках,
// поддерживая синхронизацию старых float-колонок для бесшовной миграции.
func (r *DatabaseRepository) UpdateBalance(userID int, currentDeltaCents, withdrawnDeltaCents int64) error {
	// Используем транзакцию для атомарности операции
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE user_balances 
		SET 
			current_cents   = current_cents + $1,
			withdrawn_cents = withdrawn_cents + $2,
			-- поддерживаем старые float-поля для обратной совместимости
			current   = current + ($1::numeric / 100.0),
			withdrawn = withdrawn + ($2::numeric / 100.0),
			updated_at = NOW() 
		WHERE user_id = $3`

	result, err := tx.Exec(query, currentDeltaCents, withdrawnDeltaCents, userID)
	if err != nil {
		return fmt.Errorf("не удалось обновить баланс: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось получить количество затронутых строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("баланс пользователя не найден")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("не удалось зафиксировать транзакцию: %w", err)
	}

	return nil
}

// CreateBalance создает начальный баланс для пользователя (обе схемы)
func (r *DatabaseRepository) CreateBalance(userID int) error {
	query := `
		INSERT INTO user_balances (user_id, current, withdrawn, current_cents, withdrawn_cents) 
		VALUES ($1, 0.00, 0.00, 0, 0)`

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("не удалось создать баланс: %w", err)
	}

	return nil
}
