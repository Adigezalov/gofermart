.PHONY: build run migrate migrate-status clean

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

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  build         - Собрать приложение"
	@echo "  run           - Запустить приложение"
	@echo "  migrate       - Применить миграции"
	@echo "  migrate-status - Показать статус миграций"
	@echo "  run-env       - Запустить с переменными из .env"
	@echo "  deps          - Установить зависимости"
	@echo "  clean         - Очистить собранные файлы"
	@echo "  help          - Показать эту справку"