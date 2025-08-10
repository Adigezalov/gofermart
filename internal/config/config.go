package config

import (
	"flag"
	"os"
)

// Константы для значений по умолчанию
const (
	DefaultServerAddress        = ":8080"
	DefaultDatabaseURI          = "postgres://user:password@localhost:5432/gofermart?sslmode=disable"
	DefaultAccrualSystemAddress = "http://localhost:8081"
)

// Config содержит конфигурационные параметры приложения
type Config struct {
	// ServerAddress адрес запуска HTTP-сервера
	ServerAddress string
	// DatabaseURI строка подключения к PostgreSQL
	DatabaseURI string
	// AccrualSystemAddress адрес системы расчёта начислений
	AccrualSystemAddress string
}

// NewConfig создает и инициализирует конфигурацию из аргументов командной строки и переменных окружения
func NewConfig() *Config {
	cfg := &Config{}

	// Устанавливаем значения по умолчанию
	serverAddress := DefaultServerAddress
	databaseURI := DefaultDatabaseURI
	accrualSystemAddress := DefaultAccrualSystemAddress

	// Проверяем переменные окружения
	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		serverAddress = envRunAddr
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		databaseURI = envDatabaseURI
	}
	if envAccrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualAddr != "" {
		accrualSystemAddress = envAccrualAddr
	}

	// Регистрируем флаги командной строки
	flag.StringVar(&cfg.ServerAddress, "a", serverAddress, "адрес и порт запуска сервиса")
	flag.StringVar(&cfg.DatabaseURI, "d", databaseURI, "адрес подключения к базе данных")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", accrualSystemAddress, "адрес системы расчёта начислений")

	// Разбираем флаги
	flag.Parse()

	// Валидируем и нормализуем конфигурацию
	cfg.normalize()

	return cfg
}

// normalize выполняет нормализацию и валидацию параметров конфигурации
func (c *Config) normalize() {
	// Добавляем двоеточие к адресу сервера, если его нет
	if c.ServerAddress[0] != ':' && len(c.ServerAddress) > 0 {
		if c.ServerAddress[0] >= '0' && c.ServerAddress[0] <= '9' {
			c.ServerAddress = ":" + c.ServerAddress
		}
	}
}

// GetAccrualSystemAddress возвращает адрес системы начислений
func (c *Config) GetAccrualSystemAddress() string {
	return c.AccrualSystemAddress
}
