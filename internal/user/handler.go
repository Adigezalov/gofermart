package user

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Handler обрабатывает HTTP запросы для пользователей
type Handler struct {
	service *Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Register обрабатывает POST /api/user/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	// Декодируем JSON запрос
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Регистрируем пользователя
	tokenPair, err := h.service.RegisterUser(&req)
	if err != nil {
		if strings.Contains(err.Error(), "пользователь уже существует") {
			http.Error(w, "Логин уже занят", http.StatusConflict)
			return
		}
		if strings.Contains(err.Error(), "обязателен") ||
			strings.Contains(err.Error(), "слишком длинный") ||
			strings.Contains(err.Error(), "минимум") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Отправляем только access токен в заголовке Authorization
	w.Header().Set("Authorization", "Bearer "+tokenPair.AccessToken)
	w.WriteHeader(http.StatusOK)
}

// Login обрабатывает POST /api/user/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	// Декодируем JSON запрос
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Авторизуем пользователя
	tokenPair, err := h.service.LoginUser(&req)
	if err != nil {
		if strings.Contains(err.Error(), "неверная пара логин/пароль") {
			http.Error(w, "Неверная пара логин/пароль", http.StatusUnauthorized)
			return
		}
		if strings.Contains(err.Error(), "обязателен") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Отправляем только access токен в заголовке Authorization
	w.Header().Set("Authorization", "Bearer "+tokenPair.AccessToken)
	w.WriteHeader(http.StatusOK)
}
