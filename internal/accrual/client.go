package accrual

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Client клиент для взаимодействия с внешней системой начислений
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient создает новый экземпляр Client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetOrderAccrual получает информацию о начислении для заказа
func (c *Client) GetOrderAccrual(orderNumber string) (*AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("не удалось выполнить запрос: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// Успешный ответ с данными о начислении
		var accrualResp AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return nil, fmt.Errorf("не удалось декодировать ответ: %w", err)
		}
		return &accrualResp, nil

	case http.StatusNoContent:
		// Заказ не зарегистрирован в системе расчета
		return nil, &OrderNotFoundError{OrderNumber: orderNumber}

	case http.StatusTooManyRequests:
		// Превышено количество запросов
		retryAfter := c.getRetryAfter(resp)
		return nil, &RateLimitError{
			RetryAfter: retryAfter,
			Message:    "превышено количество запросов к сервису",
		}

	case http.StatusInternalServerError:
		return nil, &ServerError{
			StatusCode: resp.StatusCode,
			Message:    "внутренняя ошибка сервера начислений",
		}

	default:
		return nil, &ServerError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("неожиданный статус ответа: %d", resp.StatusCode),
		}
	}
}

// getRetryAfter извлекает значение Retry-After из заголовков ответа
func (c *Client) getRetryAfter(resp *http.Response) time.Duration {
	retryAfterHeader := resp.Header.Get("Retry-After")
	if retryAfterHeader == "" {
		return 60 * time.Second // По умолчанию 60 секунд
	}

	// Пытаемся парсить как количество секунд
	if seconds, err := strconv.Atoi(retryAfterHeader); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Если не удалось парсить, возвращаем значение по умолчанию
	return 60 * time.Second
}

// Кастомные ошибки

// OrderNotFoundError ошибка - заказ не найден в системе начислений
type OrderNotFoundError struct {
	OrderNumber string
}

func (e *OrderNotFoundError) Error() string {
	return fmt.Sprintf("заказ %s не зарегистрирован в системе расчета", e.OrderNumber)
}

// RateLimitError ошибка превышения лимита запросов
type RateLimitError struct {
	RetryAfter time.Duration
	Message    string
}

func (e *RateLimitError) Error() string {
	return e.Message
}

// ServerError ошибка сервера
type ServerError struct {
	StatusCode int
	Message    string
}

func (e *ServerError) Error() string {
	return e.Message
}
