BEGIN;

-- Добавляем колонки в копейках для баланса
ALTER TABLE user_balances
    ADD COLUMN IF NOT EXISTS current_cents BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS withdrawn_cents BIGINT NOT NULL DEFAULT 0;

-- Инициализируем значениями из старых float-колонок (округление до копеек)
UPDATE user_balances
SET current_cents = ROUND(current * 100.0),
    withdrawn_cents = ROUND(withdrawn * 100.0)
WHERE (current_cents = 0 AND current <> 0.0)
   OR (withdrawn_cents = 0 AND withdrawn <> 0.0);

-- Добавляем колонку в копейках для списаний
ALTER TABLE withdrawals
    ADD COLUMN IF NOT EXISTS amount_cents BIGINT NOT NULL DEFAULT 0;

-- Инициализируем значениями из старой float-колонки
UPDATE withdrawals
SET amount_cents = ROUND(amount * 100.0)
WHERE amount_cents = 0 AND amount <> 0.0;

COMMIT;
