package migrations

import (
	"database/sql"
	"fmt"
	"log"
)

// Migrator управляет миграциями базы данных
type Migrator struct {
	db *sql.DB
}

// NewMigrator создает новый экземпляр Migrator
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

// RunMigrations выполняет все миграции
func (m *Migrator) RunMigrations() error {
	// Создаем таблицу для отслеживания миграций
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("не удалось создать таблицу миграций: %w", err)
	}

	// Получаем список выполненных миграций
	appliedMigrations, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("не удалось получить список выполненных миграций: %w", err)
	}

	// Получаем список файлов миграций
	migrationFiles, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("не удалось получить файлы миграций: %w", err)
	}

	// Применяем новые миграции
	for _, filename := range migrationFiles {
		if m.isMigrationApplied(filename, appliedMigrations) {
			log.Printf("Миграция %s уже применена, пропускаем", filename)
			continue
		}

		log.Printf("Применяем миграцию: %s", filename)
		if err := m.applyMigration(filename); err != nil {
			return fmt.Errorf("не удалось применить миграцию %s: %w", filename, err)
		}

		// Записываем в таблицу миграций
		if err := m.recordMigration(filename); err != nil {
			return fmt.Errorf("не удалось записать миграцию %s: %w", filename, err)
		}

		log.Printf("Миграция %s успешно применена", filename)
	}

	log.Println("Все миграции успешно применены")
	return nil
}

// createMigrationsTable создает таблицу для отслеживания миграций
func (m *Migrator) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`

	_, err := m.db.Exec(query)
	return err
}

// getAppliedMigrations получает список уже примененных миграций
func (m *Migrator) getAppliedMigrations() (map[string]bool, error) {
	query := "SELECT filename FROM schema_migrations"
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		applied[filename] = true
	}

	return applied, rows.Err()
}

// getMigrationFiles получает отсортированный список файлов миграций
func (m *Migrator) getMigrationFiles() ([]string, error) {
	// Возвращаем список встроенных миграций в правильном порядке
	files := []string{
		"001_create_users_table.sql",
		"002_create_tokens_table.sql",
		"003_create_orders_table.sql",
	}

	return files, nil
}

// isMigrationApplied проверяет, была ли миграция уже применена
func (m *Migrator) isMigrationApplied(filename string, applied map[string]bool) bool {
	return applied[filename]
}

// applyMigration применяет конкретную миграцию
func (m *Migrator) applyMigration(filename string) error {
	// Читаем содержимое файла миграции
	content, err := m.readMigrationFile(filename)
	if err != nil {
		return err
	}

	// Выполняем SQL
	_, err = m.db.Exec(content)
	return err
}

// readMigrationFile читает содержимое файла миграции
func (m *Migrator) readMigrationFile(filename string) (string, error) {
	// Используем встроенные миграции для надежности
	return m.getEmbeddedMigration(filename)
}

// getEmbeddedMigration возвращает встроенную миграцию
func (m *Migrator) getEmbeddedMigration(filename string) (string, error) {
	switch filename {
	case "001_create_users_table.sql":
		return m.getUsersTableMigration(), nil

	case "002_create_tokens_table.sql":
		return m.getTokensTableMigration(), nil

	case "003_create_orders_table.sql":
		return m.getOrdersTableMigration(), nil

	default:
		return "", fmt.Errorf("неизвестная миграция: %s", filename)
	}
}

// getUsersTableMigration возвращает SQL для создания таблицы пользователей
func (m *Migrator) getUsersTableMigration() string {
	return `-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Создание индекса для быстрого поиска по логину
CREATE INDEX IF NOT EXISTS idx_users_login ON users(login);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для автоматического обновления updated_at при изменении записи
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();`
}

// getTokensTableMigration возвращает SQL для создания таблицы токенов
func (m *Migrator) getTokensTableMigration() string {
	return `-- Создание таблицы refresh токенов
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    token VARCHAR(255) NOT NULL UNIQUE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Создание индексов
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- Триггер для автоматического обновления updated_at
DROP TRIGGER IF EXISTS update_refresh_tokens_updated_at ON refresh_tokens;
CREATE TRIGGER update_refresh_tokens_updated_at 
    BEFORE UPDATE ON refresh_tokens 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();`
}

// recordMigration записывает информацию о примененной миграции
func (m *Migrator) recordMigration(filename string) error {
	query := "INSERT INTO schema_migrations (filename) VALUES ($1)"
	_, err := m.db.Exec(query, filename)
	return err
}

// ShowStatus показывает статус миграций
func (m *Migrator) ShowStatus() error {
	// Создаем таблицу миграций если её нет
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("не удалось создать таблицу миграций: %w", err)
	}

	// Получаем список выполненных миграций
	appliedMigrations, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("не удалось получить список выполненных миграций: %w", err)
	}

	// Получаем список всех миграций
	allMigrations, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("не удалось получить список миграций: %w", err)
	}

	log.Println("Статус миграций:")
	log.Println("================")

	for _, migration := range allMigrations {
		status := "НЕ ПРИМЕНЕНА"
		if appliedMigrations[migration] {
			status = "ПРИМЕНЕНА"
		}
		log.Printf("%-30s %s", migration, status)
	}

	appliedCount := len(appliedMigrations)
	totalCount := len(allMigrations)
	log.Println("================")
	log.Printf("Применено: %d из %d миграций", appliedCount, totalCount)

	return nil
}

// getOrdersTableMigration возвращает SQL для создания таблицы заказов
func (m *Migrator) getOrdersTableMigration() string {
	return `-- Создание таблицы заказов
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    number VARCHAR(255) NOT NULL UNIQUE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'NEW',
    accrual DECIMAL(10,2) NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Создание индексов
CREATE INDEX IF NOT EXISTS idx_orders_number ON orders(number);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_uploaded_at ON orders(uploaded_at DESC);

-- Триггер для автоматического обновления updated_at
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
CREATE TRIGGER update_orders_updated_at 
    BEFORE UPDATE ON orders 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();`
}
