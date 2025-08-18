-- Откат миграции для таблицы заказов

-- Удаление триггера
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;

-- Удаление индексов
DROP INDEX IF EXISTS idx_orders_uploaded_at;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_orders_number;

-- Удаление таблицы
DROP TABLE IF EXISTS orders;