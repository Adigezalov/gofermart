.PHONY: build run migrate migrate-status clean run-accrual run-full stop-all test-system register-test-orders

# Сборка приложения
build:
	go build -o bin/gophermart ./cmd/gophermart
	go build -o bin/migrate ./cmd/migrate

# Запуск приложения
run: build
	./bin/gophermart

# Применение миграций
migrate: build
	./bin/migrate -action=up

# Проверка статуса миграций
migrate-status: build
	./bin/migrate -action=status

# Очистка
clean:
	rm -rf bin/
	rm -f .accrual.pid .gophermart.pid

# Установка зависимостей
deps:
	go mod tidy
	go mod download

# Запуск с переменными окружения из .env
run-env: build
	@if [ -f .env ]; then \
		export $$(cat .env | xargs) && ./bin/gophermart; \
	else \
		echo "Файл .env не найден, запускаем с настройками по умолчанию"; \
		./bin/gophermart; \
	fi

# Определение архитектуры для выбора правильного accrual бинарника
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Darwin)
    ifeq ($(UNAME_M),arm64)
        ACCRUAL_BINARY = cmd/accrual/accrual_darwin_arm64
    else
        ACCRUAL_BINARY = cmd/accrual/accrual_darwin_amd64
    endif
else ifeq ($(UNAME_S),Linux)
    ACCRUAL_BINARY = cmd/accrual/accrual_linux_amd64
else
    ACCRUAL_BINARY = cmd/accrual/accrual_windows_amd64
endif

# Запуск accrual сервиса
run-accrual:
	@echo "Запуск accrual сервиса ($(ACCRUAL_BINARY))..."
	@chmod +x $(ACCRUAL_BINARY)
	$(ACCRUAL_BINARY) -a localhost:8081 -d "postgresql://postgres:postgres@localhost/praktikum?sslmode=disable"

# Запуск обеих систем (gophermart + accrual)
run-full: build
	@echo "Запуск полной системы лояльности..."
	@echo "1. Запускаем accrual сервис в фоне..."
	@chmod +x $(ACCRUAL_BINARY)
	@$(ACCRUAL_BINARY) -a localhost:8081 -d "postgresql://postgres:postgres@localhost/praktikum?sslmode=disable" & \
	echo $$! > .accrual.pid
	@sleep 2
	@echo "2. Запускаем gophermart..."
	@echo "Для остановки используйте: make stop-all"
	./bin/gophermart

# Остановка всех процессов
stop-all:
	@echo "Остановка всех процессов..."
	@if [ -f .accrual.pid ]; then \
		kill `cat .accrual.pid` 2>/dev/null || true; \
		rm -f .accrual.pid; \
		echo "Accrual сервис остановлен"; \
	fi
	@pkill -f "gophermart" 2>/dev/null || true
	@echo "Все процессы остановлены"

# Быстрое тестирование системы
test-system: build
	@echo "🚀 Запуск быстрого теста системы лояльности..."
	@echo "1. Запускаем accrual сервис..."
	@chmod +x $(ACCRUAL_BINARY)
	@$(ACCRUAL_BINARY) -a localhost:8081 -d "postgresql://postgres:postgres@localhost/praktikum?sslmode=disable" & \
	echo $$! > .accrual.pid
	@sleep 3
	@echo "2. Запускаем gophermart в фоне..."
	@./bin/gophermart & echo $$! > .gophermart.pid
	@sleep 3
	@echo "3. Выполняем базовые тесты..."
	@curl -s http://localhost:8080/api/health/check > /dev/null && echo "✅ Health check работает" || echo "❌ Health check не работает"
	@curl -s http://localhost:8080/api/health/db > /dev/null && echo "✅ База данных подключена" || echo "❌ База данных недоступна"
	@curl -s http://localhost:8081 > /dev/null && echo "✅ Accrual сервис работает" || echo "❌ Accrual сервис недоступен"
	@echo "4. Останавливаем процессы..."
	@kill `cat .gophermart.pid` 2>/dev/null || true
	@kill `cat .accrual.pid` 2>/dev/null || true
	@rm -f .gophermart.pid .accrual.pid
	@echo "🎉 Быстрый тест завершен!"
	@echo ""
	@echo "Для полного тестирования:"
	@echo "  1. make run-full"
	@echo "  2. make register-test-orders (регистрация заказов в accrual)"
	@echo "  3. Откройте http-client/complete-loyalty-system.http"

# Регистрация тестовых заказов в accrual системе
register-test-orders:
	@echo "📦 Регистрация тестовых заказов в accrual системе..."
	@curl -s -X POST http://localhost:8081/api/orders -H "Content-Type: application/json" -d '{"order":"79927398713"}' > /dev/null && echo "✅ Заказ 79927398713 зарегистрирован" || echo "❌ Ошибка регистрации заказа 79927398713"
	@curl -s -X POST http://localhost:8081/api/orders -H "Content-Type: application/json" -d '{"order":"49927398716"}' > /dev/null && echo "✅ Заказ 49927398716 зарегистрирован" || echo "❌ Ошибка регистрации заказа 49927398716"
	@curl -s -X POST http://localhost:8081/api/orders -H "Content-Type: application/json" -d '{"order":"12345678903"}' > /dev/null && echo "✅ Заказ 12345678903 зарегистрирован" || echo "❌ Ошибка регистрации заказа 12345678903"
	@curl -s -X POST http://localhost:8081/api/orders -H "Content-Type: application/json" -d '{"order":"37828934750"}' > /dev/null && echo "✅ Заказ 37828934750 зарегистрирован" || echo "❌ Ошибка регистрации заказа 37828934750"
	@echo "🎉 Все тестовые заказы зарегистрированы в accrual системе!"
	@echo "Теперь accrual worker сможет их обработать."

# Запуск с переменными окружения и accrual
run-full-env: build
	@echo "Запуск полной системы с переменными из .env..."
	@chmod +x $(ACCRUAL_BINARY)
	@$(ACCRUAL_BINARY) -a localhost:8081 -d "postgresql://postgres:postgres@localhost/praktikum?sslmode=disable" & \
	echo $$! > .accrual.pid
	@sleep 2
	@if [ -f .env ]; then \
		export $(cat .env | xargs) && ./bin/gophermart; \
	else \
		echo "Файл .env не найден, запускаем с настройками по умолчанию"; \
		./bin/gophermart; \
	fi

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  build         - Собрать приложение"
	@echo "  run           - Запустить только gophermart"
	@echo "  run-accrual   - Запустить только accrual сервис"
	@echo "  run-full      - Запустить gophermart + accrual (рекомендуется)"
	@echo "  run-full-env  - Запустить полную систему с переменными из .env"
	@echo "  test-system   - Быстрый тест работоспособности системы"
	@echo "  register-test-orders - Зарегистрировать тестовые заказы в accrual"
	@echo "  stop-all      - Остановить все процессы"
	@echo "  migrate       - Применить миграции"
	@echo "  migrate-status - Показать статус миграций"
	@echo "  run-env       - Запустить gophermart с переменными из .env"
	@echo "  deps          - Установить зависимости"
	@echo "  clean         - Очистить собранные файлы"
	@echo "  help          - Показать эту справку"
	@echo ""
	@echo "Для полного тестирования системы лояльности:"
	@echo "  1. make run-full    - запустить обе системы"
	@echo "  2. Использовать http-client/complete-loyalty-system.http"
	@echo "  3. make stop-all    - остановить все процессы"