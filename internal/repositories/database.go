package repositories

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// DatabaseRepository для работы с базой данных
type DatabaseRepository struct {
	db *sql.DB
}

// NewDatabaseRepository создает новый экземпляр DatabaseRepository
func NewDatabaseRepository(dsn string) (*DatabaseRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseRepository{db: db}, nil
}

// Ping проверяет подключение к базе данных
func (dr *DatabaseRepository) Ping() error {
	return dr.db.Ping()
}

// Close закрывает подключение к базе данных
func (dr *DatabaseRepository) Close() error {
	return dr.db.Close()
}

// GetDB возвращает экземпляр базы данных для миграций
func (dr *DatabaseRepository) GetDB() *sql.DB {
	return dr.db
}
