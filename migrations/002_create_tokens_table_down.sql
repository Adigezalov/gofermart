-- Откат миграции для таблицы токенов

-- Удаление триггера
DROP TRIGGER IF EXISTS update_refresh_tokens_updated_at ON refresh_tokens;

-- Удаление индексов
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS idx_refresh_tokens_token;

-- Удаление таблицы
DROP TABLE IF EXISTS refresh_tokens;