package tokens

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Service содержит бизнес-логику для работы с токенами
type Service struct {
	repo       Repository
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewService создает новый экземпляр Service
func NewService(repo Repository, jwtSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		repo:       repo,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// GenerateTokenPair создает пару токенов для пользователя
func (s *Service) GenerateTokenPair(userID int, login string) (*TokenPair, error) {
	// Генерируем access токен
	accessToken, err := s.generateJWT(userID, login, "access", s.accessTTL)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать access токен: %w", err)
	}

	// Генерируем refresh токен
	refreshTokenString, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("не удалось создать refresh токен: %w", err)
	}

	// Сохраняем refresh токен в БД
	refreshToken := &RefreshToken{
		Token:  refreshTokenString,
		UserID: userID,
	}

	if err := s.repo.SaveRefreshToken(refreshToken); err != nil {
		return nil, fmt.Errorf("не удалось сохранить refresh токен: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
	}, nil
}

// generateJWT создает JWT токен
func (s *Service) generateJWT(userID int, login, tokenType string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"login":   login,
		"type":    tokenType,
		"exp":     time.Now().Add(ttl).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// generateRefreshToken создает случайный refresh токен
func (s *Service) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateAccessToken проверяет access токен
func (s *Service) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["type"] != "access" {
			return nil, fmt.Errorf("неверный тип токена")
		}

		return &Claims{
			UserID: int(claims["user_id"].(float64)),
			Login:  claims["login"].(string),
			Type:   claims["type"].(string),
		}, nil
	}

	return nil, fmt.Errorf("недействительный токен")
}
