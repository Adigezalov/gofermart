package withdrawal

import (
	"database/sql"
	"fmt"
)

// Repository интерфейс для работы с операциями списания в БД
type Repository interface {
	CreateWithdrawal(withdrawal *Withdrawal) error
	GetWithdrawalsByUserID(userID int) ([]*Withdrawal, error)
}

// DatabaseRepository реализация Repository для PostgreSQL
type DatabaseRepository struct {
	db *sql.DB
}

// NewDatabaseRepository создает новый экземпляр DatabaseRepository
func NewDatabaseRepository(db *sql.DB) *DatabaseRepository {
	return &DatabaseRepository{db: db}
}

// CreateWithdrawal создает новую операцию списания
func (r *DatabaseRepository) CreateWithdrawal(withdrawal *Withdrawal) error {
	query := `
		INSERT INTO withdrawals (user_id, order_number, amount) 
		VALUES ($1, $2, $3) 
		RETURNING id, processed_at`

	err := r.db.QueryRow(query, withdrawal.UserID, withdrawal.OrderNumber, withdrawal.Amount).Scan(
		&withdrawal.ID, &withdrawal.ProcessedAt,
	)
	if err != nil {
		return fmt.Errorf("не удалось создать операцию списания: %w", err)
	}

	return nil
}

// GetWithdrawalsByUserID получает все операции списания пользователя
func (r *DatabaseRepository) GetWithdrawalsByUserID(userID int) ([]*Withdrawal, error) {
	query := `
		SELECT id, user_id, order_number, amount, processed_at 
		FROM withdrawals 
		WHERE user_id = $1 
		ORDER BY processed_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить операции списания: %w", err)
	}
	defer rows.Close()

	var withdrawals []*Withdrawal
	for rows.Next() {
		withdrawal := &Withdrawal{}
		err := rows.Scan(
			&withdrawal.ID, &withdrawal.UserID, &withdrawal.OrderNumber,
			&withdrawal.Amount, &withdrawal.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("не удалось сканировать операцию списания: %w", err)
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по операциям списания: %w", err)
	}

	return withdrawals, nil
}