package tokens

// Repository интерфейс для работы с токенами в БД
type Repository interface {
	SaveRefreshToken(token *RefreshToken) error
	GetRefreshToken(token string) (*RefreshToken, error)
	DeleteRefreshToken(token string) error
	DeleteUserRefreshTokens(userID int) error
}
