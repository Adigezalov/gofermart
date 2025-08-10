package utils

import (
	"context"
	"fmt"

	"github.com/Adigezalov/gophermart/internal/middleware"
)

// GetCurrentUser извлекает информацию о текущем пользователе из контекста
func GetCurrentUser(ctx context.Context) (userID int, login string, err error) {
	userID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok {
		return 0, "", fmt.Errorf("пользователь не авторизован")
	}

	login, ok = middleware.GetUserLoginFromContext(ctx)
	if !ok {
		return 0, "", fmt.Errorf("логин пользователя не найден в контексте")
	}

	return userID, login, nil
}

// MustGetCurrentUser извлекает информацию о пользователе или паникует
func MustGetCurrentUser(ctx context.Context) (userID int, login string) {
	userID, login, err := GetCurrentUser(ctx)
	if err != nil {
		panic(err)
	}
	return userID, login
}
