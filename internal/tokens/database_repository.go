package tokens

import (
	"database/sql"
	"fmt"
)

// DatabaseRepository реализация Repository для PostgreSQL
type DatabaseRepository struct {
	db *sql.DB
}

// NewDatabaseRepository создает новый экземпляр DatabaseRepository
func NewDatabaseRepository(db *sql.DB) *DatabaseRepository {
	return &DatabaseRepository{db: db}
}

// SaveRefreshToken сохраняет refresh токен в БД
func (r *DatabaseRepository) SaveRefreshToken(token *RefreshToken) error {
	// Сначала удаляем все старые токены пользователя
	if err := r.DeleteUserRefreshTokens(token.UserID); err != nil {
		return fmt.Errorf("не удалось удалить старые токены: %w", err)
	}

	query := `
		INSERT INTO refresh_tokens (token, user_id) 
		VALUES ($1, $2) 
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query, token.Token, token.UserID).Scan(
		&token.ID, &token.CreatedAt, &token.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("не удалось сохранить refresh токен: %w", err)
	}

	return nil
}

// GetRefreshToken получает refresh токен из БД
func (r *DatabaseRepository) GetRefreshToken(tokenString string) (*RefreshToken, error) {
	token := &RefreshToken{}
	query := `
		SELECT id, token, user_id, created_at, updated_at 
		FROM refresh_tokens 
		WHERE token = $1`

	err := r.db.QueryRow(query, tokenString).Scan(
		&token.ID, &token.Token, &token.UserID, &token.CreatedAt, &token.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("токен не найден")
		}
		return nil, fmt.Errorf("не удалось получить refresh токен: %w", err)
	}

	return token, nil
}

// DeleteRefreshToken удаляет refresh токен из БД
func (r *DatabaseRepository) DeleteRefreshToken(tokenString string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`

	result, err := r.db.Exec(query, tokenString)
	if err != nil {
		return fmt.Errorf("не удалось удалить refresh токен: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось получить количество затронутых строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("токен не найден")
	}

	return nil
}

// DeleteUserRefreshTokens удаляет все refresh токены пользователя
func (r *DatabaseRepository) DeleteUserRefreshTokens(userID int) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("не удалось удалить refresh токены пользователя: %w", err)
	}

	return nil
}
