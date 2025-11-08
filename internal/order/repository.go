package order

import (
	"database/sql"
	"fmt"
	"strings"
)

// Repository интерфейс для работы с заказами в БД
type Repository interface {
	CreateOrder(order *Order) error
	GetOrderByNumber(number string) (*Order, error)
	GetOrdersByUserID(userID int) ([]*Order, error)
	GetOrdersByStatus(statuses []OrderStatus) ([]*Order, error)
	UpdateOrderStatus(number string, status OrderStatus, accrual *float64) error
}

// DatabaseRepository реализация Repository для PostgreSQL
type DatabaseRepository struct {
	db *sql.DB
}

// NewDatabaseRepository создает новый экземпляр DatabaseRepository
func NewDatabaseRepository(db *sql.DB) *DatabaseRepository {
	return &DatabaseRepository{db: db}
}

// CreateOrder создает новый заказ
func (r *DatabaseRepository) CreateOrder(order *Order) error {
	query := `
		INSERT INTO orders (number, user_id, status) 
		VALUES ($1, $2, $3) 
		RETURNING id, uploaded_at, updated_at`

	err := r.db.QueryRow(query, order.Number, order.UserID, order.Status).Scan(
		&order.ID, &order.UploadedAt, &order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("не удалось создать заказ: %w", err)
	}

	return nil
}

// GetOrderByNumber получает заказ по номеру
func (r *DatabaseRepository) GetOrderByNumber(number string) (*Order, error) {
	order := &Order{}
	query := `
		SELECT id, number, user_id, status, accrual, uploaded_at, updated_at 
		FROM orders 
		WHERE number = $1`

	err := r.db.QueryRow(query, number).Scan(
		&order.ID, &order.Number, &order.UserID, &order.Status,
		&order.Accrual, &order.UploadedAt, &order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("заказ не найден")
		}
		return nil, fmt.Errorf("не удалось получить заказ: %w", err)
	}

	return order, nil
}

// GetOrdersByUserID получает все заказы пользователя
func (r *DatabaseRepository) GetOrdersByUserID(userID int) ([]*Order, error) {
	query := `
		SELECT id, number, user_id, status, accrual, uploaded_at, updated_at 
		FROM orders 
		WHERE user_id = $1 
		ORDER BY uploaded_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить заказы пользователя: %w", err)
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		order := &Order{}
		err := rows.Scan(
			&order.ID, &order.Number, &order.UserID, &order.Status,
			&order.Accrual, &order.UploadedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("не удалось сканировать заказ: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по заказам: %w", err)
	}

	return orders, nil
}

// GetOrdersByStatus получает заказы по статусам
func (r *DatabaseRepository) GetOrdersByStatus(statuses []OrderStatus) ([]*Order, error) {
	if len(statuses) == 0 {
		return []*Order{}, nil
	}

	// Создаем плейсхолдеры для IN clause
	placeholders := make([]string, len(statuses))
	args := make([]interface{}, len(statuses))
	for i, status := range statuses {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = status
	}

	query := fmt.Sprintf(`
		SELECT id, number, user_id, status, accrual, uploaded_at, updated_at 
		FROM orders 
		WHERE status IN (%s) 
		ORDER BY uploaded_at ASC`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить заказы по статусу: %w", err)
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		order := &Order{}
		err := rows.Scan(
			&order.ID, &order.Number, &order.UserID, &order.Status,
			&order.Accrual, &order.UploadedAt, &order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("не удалось сканировать заказ: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по заказам: %w", err)
	}

	return orders, nil
}

// UpdateOrderStatus обновляет статус и начисление заказа
func (r *DatabaseRepository) UpdateOrderStatus(number string, status OrderStatus, accrual *float64) error {
	query := `
		UPDATE orders 
		SET status = $1, accrual = $2, updated_at = NOW() 
		WHERE number = $3`

	result, err := r.db.Exec(query, status, accrual, number)
	if err != nil {
		return fmt.Errorf("не удалось обновить статус заказа: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось получить количество затронутых строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("заказ не найден")
	}

	return nil
}
