package balance

import (
	"net/http"

	"github.com/Adigezalov/gophermart/internal/utils"
)

// Handler обрабатывает HTTP запросы для балансов
type Handler struct {
	service *Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetBalance обрабатывает GET /api/user/balance
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из контекста (установлен middleware авторизации)
	userID, _, err := utils.GetCurrentUser(r.Context())
	if err != nil {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	// Получаем баланс пользователя через сервис
	balance, err := h.service.GetUserBalance(userID)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Возвращаем баланс в формате JSON
	utils.WriteJSON(w, balance)
}