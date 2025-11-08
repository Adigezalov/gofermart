package user

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Adigezalov/gophermart/internal/balance"
	"github.com/Adigezalov/gophermart/internal/tokens"
	"golang.org/x/crypto/bcrypt"
)

// Service содержит бизнес-логику для пользователей
type Service struct {
	repo           Repository
	tokenService   *tokens.Service
	balanceService *balance.Service
}

// NewService создает новый экземпляр Service
func NewService(repo Repository, tokenService *tokens.Service, balanceService *balance.Service) *Service {
	return &Service{
		repo:           repo,
		tokenService:   tokenService,
		balanceService: balanceService,
	}
}

// RegisterUser регистрирует нового пользователя
func (s *Service) RegisterUser(req *RegisterRequest) (*tokens.TokenPair, error) {
	// Валидация входных данных
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// Проверяем, не существует ли уже пользователь с таким логином
	existingUser, err := s.repo.GetUserByLogin(req.Login)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("пользователь уже существует")
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("не удалось захешировать пароль: %w", err)
	}

	// Создаем пользователя
	user := &User{
		Login:        req.Login,
		PasswordHash: string(hashedPassword),
	}

	if err := s.repo.CreateUser(user); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, fmt.Errorf("пользователь уже существует")
		}
		return nil, fmt.Errorf("не удалось создать пользователя: %w", err)
	}

	// Создаем начальный баланс для пользователя
	if s.balanceService != nil {
		if err := s.balanceService.EnsureBalance(user.ID); err != nil {
			// Логируем предупреждение, но не прерываем регистрацию
			log.Printf("Предупреждение: не удалось создать баланс для пользователя %d: %v", user.ID, err)
		}
	}

	// Генерируем токены
	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Login)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать токены: %w", err)
	}

	return tokenPair, nil
}

// validateRegisterRequest валидирует запрос на регистрацию
func (s *Service) validateRegisterRequest(req *RegisterRequest) error {
	if req == nil {
		return errors.New("запрос обязателен")
	}

	if strings.TrimSpace(req.Login) == "" {
		return errors.New("логин обязателен")
	}

	if strings.TrimSpace(req.Password) == "" {
		return errors.New("пароль обязателен")
	}

	if len(req.Login) > 255 {
		return errors.New("логин слишком длинный")
	}

	if len(req.Password) < 6 {
		return errors.New("пароль должен содержать минимум 6 символов")
	}

	return nil
}

// LoginUser авторизует пользователя
func (s *Service) LoginUser(req *LoginRequest) (*tokens.TokenPair, error) {
	// Валидация входных данных
	if err := s.validateLoginRequest(req); err != nil {
		return nil, err
	}

	// Получаем пользователя по логину
	user, err := s.repo.GetUserByLogin(req.Login)
	if err != nil {
		return nil, fmt.Errorf("неверная пара логин/пароль")
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("неверная пара логин/пароль")
	}

	// Генерируем новые токены (старые refresh токены автоматически удалятся)
	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.Login)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать токены: %w", err)
	}

	return tokenPair, nil
}

// validateLoginRequest валидирует запрос на авторизацию
func (s *Service) validateLoginRequest(req *LoginRequest) error {
	if req == nil {
		return errors.New("запрос обязателен")
	}

	if strings.TrimSpace(req.Login) == "" {
		return errors.New("логин обязателен")
	}

	if strings.TrimSpace(req.Password) == "" {
		return errors.New("пароль обязателен")
	}

	return nil
}
