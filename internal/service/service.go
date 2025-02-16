package service

type Repository interface {
	CoinTransferRepository
	PurchaseRepository
	AuthRepository
	UserRepository
}

type Service struct {
	*AuthService
	*CoinTransferService
	*PurchaseService
	*UserService
}

func NewService(repo Repository, jwtSecret string) *Service {
	return &Service{
		AuthService:         NewAuthService(repo, jwtSecret),
		CoinTransferService: NewCoinTransferService(repo),
		PurchaseService:     NewPurchaseService(repo),
		UserService:         NewUserService(repo),
	}
}
