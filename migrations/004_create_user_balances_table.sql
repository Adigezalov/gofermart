-- Создание таблицы балансов пользователей
CREATE TABLE IF NOT EXISTS user_balances (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    current DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    withdrawn DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id)
);

-- Создание индексов
CREATE INDEX IF NOT EXISTS idx_user_balances_user_id ON user_balances(user_id);

-- Триггер для автоматического обновления updated_at
DROP TRIGGER IF EXISTS update_user_balances_updated_at ON user_balances;
CREATE TRIGGER update_user_balances_updated_at 
    BEFORE UPDATE ON user_balances 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Создание балансов для существующих пользователей
INSERT INTO user_balances (user_id, current, withdrawn)
SELECT id, 0.00, 0.00 
FROM users 
WHERE id NOT IN (SELECT user_id FROM user_balances);