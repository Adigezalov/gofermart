package user

import (
	"database/sql"
	"fmt"
)

// Repository интерфейс для работы с пользователями в БД
type Repository interface {
	CreateUser(user *User) error
	GetUserByLogin(login string) (*User, error)
	GetUserByID(id int) (*User, error)
}

// DatabaseRepository реализация Repository для PostgreSQL
type DatabaseRepository struct {
	db *sql.DB
}

// NewDatabaseRepository создает новый экземпляр DatabaseRepository
func NewDatabaseRepository(db *sql.DB) *DatabaseRepository {
	return &DatabaseRepository{db: db}
}

// CreateUser создает нового пользователя
func (r *DatabaseRepository) CreateUser(user *User) error {
	query := `
		INSERT INTO users (login, password_hash) 
		VALUES ($1, $2) 
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query, user.Login, user.PasswordHash).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("не удалось создать пользователя: %w", err)
	}

	return nil
}

// GetUserByLogin получает пользователя по логину
func (r *DatabaseRepository) GetUserByLogin(login string) (*User, error) {
	user := &User{}
	query := `
		SELECT id, login, password_hash, created_at, updated_at 
		FROM users 
		WHERE login = $1`

	err := r.db.QueryRow(query, login).Scan(
		&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("не удалось получить пользователя: %w", err)
	}

	return user, nil
}

// GetUserByID получает пользователя по ID
func (r *DatabaseRepository) GetUserByID(id int) (*User, error) {
	user := &User{}
	query := `
		SELECT id, login, password_hash, created_at, updated_at 
		FROM users 
		WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("пользователь не найден")
		}
		return nil, fmt.Errorf("не удалось получить пользователя: %w", err)
	}

	return user, nil
}
