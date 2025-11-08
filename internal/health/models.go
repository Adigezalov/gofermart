package health

// Response представляет ответ для health check
type Response struct {
	Status string `json:"status"`
}

// DatabaseResponse представляет ответ для проверки БД
type DatabaseResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}
