package service

import "context"

type PurchaseRepository interface {
	Buy(ctx context.Context, userID, item string) error
}

type PurchaseService struct {
	repo PurchaseRepository
}

func NewPurchaseService(repo PurchaseRepository) *PurchaseService {
	return &PurchaseService{repo: repo}
}

func (s *PurchaseService) BuyItem(ctx context.Context, userID, item string) error {
	return s.repo.Buy(ctx, userID, item)
}
