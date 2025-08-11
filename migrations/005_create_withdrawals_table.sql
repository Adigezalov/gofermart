-- Создание таблицы списаний
CREATE TABLE IF NOT EXISTS withdrawals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_number VARCHAR(255) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Создание индексов для производительности
CREATE INDEX IF NOT EXISTS idx_withdrawals_user_id ON withdrawals(user_id);
CREATE INDEX IF NOT EXISTS idx_withdrawals_processed_at ON withdrawals(processed_at DESC);
CREATE INDEX IF NOT EXISTS idx_withdrawals_order_number ON withdrawals(order_number);