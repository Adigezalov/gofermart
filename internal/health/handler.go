package health

import (
	"encoding/json"
	"net/http"
)

// Handler обрабатывает запросы для проверки состояния сервера
type Handler struct {
	service *Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Check обрабатывает GET /api/health/check
func (h *Handler) Check(w http.ResponseWriter, r *http.Request) {
	response := h.service.GetHealthStatus()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}

// CheckDatabase обрабатывает GET /api/health/db
func (h *Handler) CheckDatabase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := h.service.GetDatabaseHealthStatus()

	statusCode := http.StatusOK
	if response.Status == "error" {
		statusCode = http.StatusServiceUnavailable
	}

	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

// CheckAuth обрабатывает GET /api/health/auth - защищенный healthcheck
func (h *Handler) CheckAuth(w http.ResponseWriter, r *http.Request) {
	// Этот метод будет вызван только для авторизованных пользователей
	// благодаря middleware
	response := &Response{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}
