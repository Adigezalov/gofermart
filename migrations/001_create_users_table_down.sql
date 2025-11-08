-- Откат миграции для таблицы пользователей

-- Удаление триггера
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Удаление функции
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаление индекса
DROP INDEX IF EXISTS idx_users_login;

-- Удаление таблицы
DROP TABLE IF EXISTS users;