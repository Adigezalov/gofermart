package accrual

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Adigezalov/gophermart/internal/balance"
	"github.com/Adigezalov/gophermart/internal/order"
)

// Worker фоновый процесс для обработки заказов и синхронизации с системой начислений
type Worker struct {
	client         *Client
	orderRepo      order.Repository
	balanceService *balance.Service
	interval       time.Duration
}

// NewWorker создает новый экземпляр Worker
func NewWorker(client *Client, orderRepo order.Repository, balanceService *balance.Service) *Worker {
	return &Worker{
		client:         client,
		orderRepo:      orderRepo,
		balanceService: balanceService,
		interval:       30 * time.Second, // Проверяем каждые 30 секунд
	}
}

// Start запускает фоновый процесс обработки заказов
func (w *Worker) Start(ctx context.Context) {
	log.Println("Запуск фонового процесса обработки заказов...")

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Выполняем первую обработку сразу
	w.processOrders()

	for {
		select {
		case <-ctx.Done():
			log.Println("Остановка фонового процесса обработки заказов...")
			return
		case <-ticker.C:
			w.processOrders()
		}
	}
}

// processOrders обрабатывает заказы, требующие проверки статуса
func (w *Worker) processOrders() {
	// Получаем заказы со статусами NEW и PROCESSING
	orders, err := w.getOrdersForProcessing()
	if err != nil {
		log.Printf("Ошибка при получении заказов для обработки: %v", err)
		return
	}

	if len(orders) == 0 {
		return // Нет заказов для обработки
	}

	log.Printf("Обрабатываем %d заказов...", len(orders))

	for _, ord := range orders {
		if err := w.processOrder(ord); err != nil {
			log.Printf("Ошибка при обработке заказа %s: %v", ord.Number, err)

			// При ошибке rate limiting делаем паузу
			if rateLimitErr, ok := err.(*RateLimitError); ok {
				log.Printf("Rate limit достигнут, ждем %v", rateLimitErr.RetryAfter)
				time.Sleep(rateLimitErr.RetryAfter)
				return // Прерываем обработку до следующего цикла
			}
		}
	}
}

// getOrdersForProcessing получает заказы, которые нужно проверить
func (w *Worker) getOrdersForProcessing() ([]*order.Order, error) {
	// Получаем заказы со статусами NEW и PROCESSING
	statuses := []order.OrderStatus{order.StatusNew, order.StatusProcessing}
	return w.orderRepo.GetOrdersByStatus(statuses)
}

// processOrder обрабатывает один заказ
func (w *Worker) processOrder(ord *order.Order) error {
	// Запрашиваем информацию о начислении
	accrualResp, err := w.client.GetOrderAccrual(ord.Number)
	if err != nil {
		// Обрабатываем специальные ошибки
		switch e := err.(type) {
		case *OrderNotFoundError:
			// Заказ не найден в системе начислений - это нормально, просто ждем
			log.Printf("Заказ %s еще не зарегистрирован в системе начислений", ord.Number)
			return nil

		case *RateLimitError:
			// Превышен лимит запросов - возвращаем ошибку для обработки выше
			return e

		case *ServerError:
			// Ошибка сервера - логируем и продолжаем
			log.Printf("Ошибка сервера начислений для заказа %s: %v", ord.Number, e)
			return nil

		default:
			// Другие ошибки (сеть, таймаут и т.д.)
			return fmt.Errorf("не удалось получить информацию о начислении: %w", err)
		}
	}

	// Валидируем ответ
	if err := accrualResp.Validate(); err != nil {
		log.Printf("Невалидный ответ для заказа %s: %v", ord.Number, err)
		return nil
	}

	// Обрабатываем ответ в зависимости от статуса
	return w.handleAccrualResponse(ord, accrualResp)
}

// handleAccrualResponse обрабатывает ответ от системы начислений
func (w *Worker) handleAccrualResponse(ord *order.Order, resp *AccrualResponse) error {
	switch resp.Status {
	case StatusRegistered, StatusProcessing:
		// Заказ в обработке - обновляем статус если изменился
		if ord.Status != order.OrderStatus(resp.Status) {
			log.Printf("Обновляем статус заказа %s: %s -> %s", ord.Number, ord.Status, resp.Status)
			return w.orderRepo.UpdateOrderStatus(ord.Number, order.OrderStatus(resp.Status), nil)
		}
		return nil

	case StatusProcessed:
		// Заказ обработан - обновляем статус и начисляем баллы
		log.Printf("Заказ %s обработан, начисляем %.2f баллов пользователю %d",
			ord.Number, *resp.Accrual, ord.UserID)

		// Обновляем статус заказа
		if err := w.orderRepo.UpdateOrderStatus(ord.Number, order.StatusProcessed, resp.Accrual); err != nil {
			return fmt.Errorf("не удалось обновить статус заказа: %w", err)
		}

		// Начисляем баллы пользователю
		if err := w.balanceService.AddPoints(ord.UserID, *resp.Accrual); err != nil {
			log.Printf("КРИТИЧЕСКАЯ ОШИБКА: не удалось начислить баллы пользователю %d за заказ %s: %v",
				ord.UserID, ord.Number, err)
			// Не возвращаем ошибку, чтобы не блокировать обработку других заказов
		}

		return nil

	case StatusInvalid:
		// Заказ отклонен - обновляем статус
		log.Printf("Заказ %s отклонен системой начислений", ord.Number)
		return w.orderRepo.UpdateOrderStatus(ord.Number, order.StatusInvalid, nil)

	default:
		log.Printf("Неизвестный статус %s для заказа %s", resp.Status, ord.Number)
		return nil
	}
}
