package order

import (
	"io"
	"net/http"
	"strings"

	"github.com/Adigezalov/gophermart/internal/utils"
)

// Handler обрабатывает HTTP запросы для заказов
type Handler struct {
	service *Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// SubmitOrder обрабатывает POST /api/user/orders
func (h *Handler) SubmitOrder(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из контекста (установлен middleware авторизации)
	userID, _, err := utils.GetCurrentUser(r.Context())
	if err != nil {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	// Читаем номер заказа из тела запроса (text/plain)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		http.Error(w, "Номер заказа не может быть пустым", http.StatusBadRequest)
		return
	}

	// Загружаем заказ через сервис
	err = h.service.SubmitOrder(userID, orderNumber)
	if err != nil {
		// Обрабатываем кастомные ошибки
		switch e := err.(type) {
		case *OrderAlreadyExistsError:
			// 200 - заказ уже был загружен этим пользователем
			w.WriteHeader(http.StatusOK)
			return

		case *OrderAcceptedError:
			// 202 - новый заказ принят в обработку
			w.WriteHeader(http.StatusAccepted)
			return

		case *OrderValidationError:
			// 400/422 - ошибки валидации
			http.Error(w, e.Message, e.Code)
			return

		case *OrderConflictError:
			// 409 - заказ загружен другим пользователем
			http.Error(w, e.Message, http.StatusConflict)
			return

		default:
			// 500 - внутренняя ошибка сервера
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
	}

	// Этот код не должен выполняться, но на всякий случай
	w.WriteHeader(http.StatusAccepted)
}

// GetOrders обрабатывает GET /api/user/orders
func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из контекста (установлен middleware авторизации)
	userID, _, err := utils.GetCurrentUser(r.Context())
	if err != nil {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	// Получаем заказы пользователя через сервис
	orders, err := h.service.GetUserOrders(userID)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Если у пользователя нет заказов
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Возвращаем заказы в формате JSON
	utils.WriteJSON(w, orders)
}
