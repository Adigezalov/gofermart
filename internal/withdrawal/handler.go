package withdrawal

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Adigezalov/gophermart/internal/utils"
)

// Handler обрабатывает HTTP запросы для операций списания
type Handler struct {
	service *Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// WithdrawPoints обрабатывает POST /api/user/balance/withdraw
func (h *Handler) WithdrawPoints(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из контекста (установлен middleware авторизации)
	userID, _, err := utils.GetCurrentUser(r.Context())
	if err != nil {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Парсим JSON
	var request WithdrawRequest
	if err := json.Unmarshal(body, &request); err != nil {
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Выполняем списание через сервис
	err = h.service.WithdrawPoints(userID, request.OrderNumber, request.Amount)
	if err != nil {
		// Обрабатываем кастомные ошибки
		switch e := err.(type) {
		case *ValidationError:
			http.Error(w, e.Message, e.Code)
			return

		case *InsufficientFundsError:
			http.Error(w, "на счету недостаточно средств", http.StatusPaymentRequired)
			return

		default:
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
	}

	// Успешное списание
	w.WriteHeader(http.StatusOK)
}

// GetWithdrawals обрабатывает GET /api/user/withdrawals
func (h *Handler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	// Получаем ID пользователя из контекста (установлен middleware авторизации)
	userID, _, err := utils.GetCurrentUser(r.Context())
	if err != nil {
		http.Error(w, "Пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	// Получаем историю списаний через сервис
	withdrawals, err := h.service.GetUserWithdrawals(userID)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Если нет операций списания
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Готовим ответ: конвертируем суммы из копеек в рубли и используем ожидаемую схему
	type item struct {
		Order       string  `json:"order"`
		Sum         float64 `json:"sum"`
		ProcessedAt string  `json:"processed_at"`
	}
	resp := make([]item, 0, len(withdrawals))
	for _, wd := range withdrawals {
		resp = append(resp, item{
			Order:       wd.OrderNumber,
			Sum:         float64(wd.AmountCents) / 100.0,
			ProcessedAt: wd.ProcessedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Устанавливаем заголовок Content-Type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Возвращаем историю списаний в формате JSON
	utils.WriteJSON(w, resp)
}
