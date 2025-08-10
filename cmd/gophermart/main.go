package main

import (
	"log"
	"net/http"

	"github.com/Adigezalov/gophermart/internal/config"
	"github.com/Adigezalov/gophermart/internal/health"
	"github.com/Adigezalov/gophermart/internal/migrations"
	"github.com/Adigezalov/gophermart/internal/repositories"
	"github.com/Adigezalov/gophermart/internal/tokens"
	"github.com/Adigezalov/gophermart/internal/user"
	"github.com/gorilla/mux"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.NewConfig()

	// Создаем подключение к базе данных
	dbRepo, err := repositories.NewDatabaseRepository(cfg.DatabaseURI)
	if err != nil {
		log.Printf("Предупреждение: Не удалось подключиться к базе данных: %v", err)
		dbRepo = nil // Продолжаем работу без БД
	}

	// Создаем роутер
	router := mux.NewRouter()

	// Создаем health домен
	var healthRepo health.Repository
	if dbRepo != nil {
		healthRepo = dbRepo
	}
	healthService := health.NewService(healthRepo)
	healthHandler := health.NewHandler(healthService)

	// Настраиваем маршруты
	api := router.PathPrefix("/api").Subrouter()

	// Health check маршруты
	api.HandleFunc("/health/check", healthHandler.Check).Methods("GET")
	api.HandleFunc("/health/db", healthHandler.CheckDatabase).Methods("GET")

	// Создаем сервисы только если есть подключение к БД
	if dbRepo != nil {
		// Применяем миграции
		log.Println("Применяем миграции базы данных...")
		migrator := migrations.NewMigrator(dbRepo.GetDB())
		if err := migrator.RunMigrations(); err != nil {
			log.Fatalf("Ошибка при применении миграций: %v", err)
		}

		// Создаем репозитории
		userRepo := user.NewDatabaseRepository(dbRepo.GetDB())
		tokenRepo := tokens.NewDatabaseRepository(dbRepo.GetDB())

		// Создаем сервисы
		tokenService := tokens.NewService(tokenRepo, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
		userService := user.NewService(userRepo, tokenService)

		// Создаем handlers
		userHandler := user.NewHandler(userService)

		// Пользовательские маршруты
		userRoutes := api.PathPrefix("/user").Subrouter()
		userRoutes.HandleFunc("/register", userHandler.Register).Methods("POST")
	}

	// Запускаем сервер
	log.Printf("Сервер запущен на %s", cfg.ServerAddress)
	log.Printf("База данных: %s", cfg.DatabaseURI)
	log.Printf("Система начислений: %s", cfg.AccrualSystemAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, router))
}
