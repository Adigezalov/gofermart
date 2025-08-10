package health

// Service содержит бизнес-логику для health checks
type Service struct {
	repo Repository
}

// NewService создает новый экземпляр Service
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// GetHealthStatus возвращает общий статус здоровья сервиса
func (s *Service) GetHealthStatus() *Response {
	return &Response{
		Status: "ok",
	}
}

// GetDatabaseHealthStatus проверяет состояние базы данных
func (s *Service) GetDatabaseHealthStatus() *DatabaseResponse {
	if s.repo == nil {
		return &DatabaseResponse{
			Status:   "error",
			Database: "not configured",
		}
	}

	if err := s.repo.Ping(); err != nil {
		return &DatabaseResponse{
			Status:   "error",
			Database: "connection failed",
		}
	}

	return &DatabaseResponse{
		Status:   "ok",
		Database: "connected",
	}
}
