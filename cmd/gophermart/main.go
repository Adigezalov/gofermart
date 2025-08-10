package main

import (
	"github.com/Adigezalov/gophermart/internal/config"
	"github.com/Adigezalov/gophermart/internal/health"
	"github.com/Adigezalov/gophermart/internal/repositories"

	"github.com/gorilla/mux"

	"log"
	"net/http"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.NewConfig()

	// Создаем подключение к базе данных
	dbRepo, err := repositories.NewDatabaseRepository(cfg.DatabaseURI)
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v", err)
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

	// Запускаем сервер
	log.Printf("Сервер запущен на %s", cfg.ServerAddress)
	log.Printf("База данных: %s", cfg.DatabaseURI)
	log.Printf("Система начислений: %s", cfg.AccrualSystemAddress)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, router))
}
