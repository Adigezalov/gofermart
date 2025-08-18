package health

// Repository интерфейс для проверки состояния хранилища данных
type Repository interface {
	Ping() error
}
