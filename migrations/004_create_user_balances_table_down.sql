-- Удаление таблицы балансов пользователей
DROP TRIGGER IF EXISTS update_user_balances_updated_at ON user_balances;
DROP TABLE IF EXISTS user_balances;