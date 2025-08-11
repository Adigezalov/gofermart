package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Adigezalov/gophermart/internal/accrual"
	"github.com/Adigezalov/gophermart/internal/balance"
	"github.com/Adigezalov/gophermart/internal/config"
	"github.com/Adigezalov/gophermart/internal/health"
	"github.com/Adigezalov/gophermart/internal/middleware"
	"github.com/Adigezalov/gophermart/internal/migrations"
	"github.com/Adigezalov/gophermart/internal/order"
	"github.com/Adigezalov/gophermart/internal/repositories"
	"github.com/Adigezalov/gophermart/internal/tokens"
	"github.com/Adigezalov/gophermart/internal/user"
	"github.com/Adigezalov/gophermart/internal/withdrawal"
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

	// Контекст для управления жизненным циклом приложения
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
		balanceRepo := balance.NewDatabaseRepository(dbRepo.GetDB())
		withdrawalRepo := withdrawal.NewDatabaseRepository(dbRepo.GetDB())

		// Создаем сервисы
		tokenService := tokens.NewService(tokenRepo, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
		balanceService := balance.NewService(balanceRepo)
		userService := user.NewService(userRepo, tokenService, balanceService)
		orderService := order.NewService(orderRepo)
		withdrawalService := withdrawal.NewService(withdrawalRepo, balanceService)

		// Создаем middleware
		authMiddleware := middleware.NewAuthMiddleware(tokenService)

		// Создаем handlers
		userHandler := user.NewHandler(userService)
		orderHandler := order.NewHandler(orderService)
		balanceHandler := balance.NewHandler(balanceService)
		withdrawalHandler := withdrawal.NewHandler(withdrawalService)

		// Создаем accrual клиент и worker
		accrualClient := accrual.NewClient(cfg.AccrualSystemAddress)
		accrualWorker := accrual.NewWorker(accrualClient, orderRepo, balanceService)

		// Запускаем accrual worker в отдельной горутине
		go accrualWorker.Start(ctx)

		// Пользовательские маршруты
		userRoutes := api.PathPrefix("/user").Subrouter()
		userRoutes.HandleFunc("/register", userHandler.Register).Methods("POST")
		userRoutes.HandleFunc("/login", userHandler.Login).Methods("POST")

		// Защищенные маршруты для заказов
		userRoutes.HandleFunc("/orders", authMiddleware.RequireAuth(orderHandler.SubmitOrder)).Methods("POST")
		userRoutes.HandleFunc("/orders", authMiddleware.RequireAuth(orderHandler.GetOrders)).Methods("GET")

		// Защищенные маршруты для баланса
		userRoutes.HandleFunc("/balance", authMiddleware.RequireAuth(balanceHandler.GetBalance)).Methods("GET")
		userRoutes.HandleFunc("/balance/withdraw", authMiddleware.RequireAuth(withdrawalHandler.WithdrawPoints)).Methods("POST")

		// Защищенные маршруты для операций списания
		userRoutes.HandleFunc("/withdrawals", authMiddleware.RequireAuth(withdrawalHandler.GetWithdrawals)).Methods("GET")

		// Защищенные маршруты
		api.HandleFunc("/health/auth", authMiddleware.RequireAuth(healthHandler.CheckAuth)).Methods("GET")
		log.Println("Зарегистрированы пользовательские и защищенные маршруты")
	}



	// Настраиваем graceful shutdown
	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: router,
	}

	// Канал для получения сигналов ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Сервер запущен на %s", cfg.ServerAddress)
		log.Printf("База данных: %s", cfg.DatabaseURI)
		log.Printf("Система начислений: %s", cfg.AccrualSystemAddress)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Ждем сигнал завершения
	<-sigChan
	log.Println("Получен сигнал завершения, останавливаем сервер...")

	// Останавливаем accrual worker
	cancel()

	// Останавливаем HTTP сервер
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Ошибка при остановке сервера: %v", err)
	} else {
		log.Println("Сервер успешно остановлен")
	}
}
