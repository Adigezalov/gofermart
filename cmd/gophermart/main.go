package main

import (
	"log"
	"net/http"

	"github.com/Adigezalov/gophermart/internal/config"
	"github.com/Adigezalov/gophermart/internal/health"
	"github.com/Adigezalov/gophermart/internal/middleware"
	"github.com/Adigezalov/gophermart/internal/migrations"
	"github.com/Adigezalov/gophermart/internal/order"
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

	// Создаем logging middleware
	loggingMiddleware := middleware.NewLoggingMiddleware()

	// Применяем logging middleware ко всем запросам
	router.Use(loggingMiddleware.LogRequest)

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
	log.Println("Зарегистрированы публичные health check маршруты")

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
		orderRepo := order.NewDatabaseRepository(dbRepo.GetDB())

		// Создаем сервисы
		tokenService := tokens.NewService(tokenRepo, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
		userService := user.NewService(userRepo, tokenService)
		orderService := order.NewService(orderRepo)

		// Создаем middleware
		authMiddleware := middleware.NewAuthMiddleware(tokenService)

		// Создаем handlers
		userHandler := user.NewHandler(userService)
		orderHandler := order.NewHandler(orderService)

		// Пользовательские маршруты
		userRoutes := api.PathPrefix("/user").Subrouter()
		userRoutes.HandleFunc("/register", userHandler.Register).Methods("POST")
		userRoutes.HandleFunc("/login", userHandler.Login).Methods("POST")

		// Защищенные маршруты для заказов
		userRoutes.HandleFunc("/orders", authMiddleware.RequireAuth(orderHandler.SubmitOrder)).Methods("POST")
		userRoutes.HandleFunc("/orders", authMiddleware.RequireAuth(orderHandler.GetOrders)).Methods("GET")

		// Защищенные маршруты
		api.HandleFunc("/health/auth", authMiddleware.RequireAuth(healthHandler.CheckAuth)).Methods("GET")
		log.Println("Зарегистрированы пользовательские и защищенные маршруты")
	}

	// Запускаем сервер
	log.Printf("Сервер запущен на %s", cfg.ServerAddress)
	log.Printf("База данных: %s", cfg.DatabaseURI)
	log.Printf("Система начислений: %s", cfg.AccrualSystemAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, router))
}
