package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Adigezalov/gophermart/internal/tokens"
)

// UserContextKey ключ для хранения данных пользователя в контексте
type UserContextKey string

const (
	UserIDKey    UserContextKey = "user_id"
	UserLoginKey UserContextKey = "user_login"
)

// AuthMiddleware middleware для проверки авторизации
type AuthMiddleware struct {
	tokenService *tokens.Service
}

// NewAuthMiddleware создает новый экземпляр AuthMiddleware
func NewAuthMiddleware(tokenService *tokens.Service) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// RequireAuth проверяет наличие и валидность access токена
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Токен авторизации отсутствует", http.StatusUnauthorized)
			return
		}

		fmt.Println(authHeader)

		// Проверяем формат Bearer токена
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Неверный формат токена авторизации", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Валидируем токен
		claims, err := m.tokenService.ValidateAccessToken(tokenString)
		if err != nil {
			http.Error(w, "Недействительный токен авторизации", http.StatusUnauthorized)
			return
		}

		// Добавляем данные пользователя в контекст
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserLoginKey, claims.Login)
		r = r.WithContext(ctx)

		// Передаем управление следующему обработчику
		next(w, r)
	}
}

// GetUserIDFromContext извлекает ID пользователя из контекста
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDKey).(int)
	return userID, ok
}

// GetUserLoginFromContext извлекает логин пользователя из контекста
func GetUserLoginFromContext(ctx context.Context) (string, bool) {
	login, ok := ctx.Value(UserLoginKey).(string)
	return login, ok
}
