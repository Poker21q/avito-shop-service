package service

import "context"

type CoinTransferRepository interface {
	SendCoins(ctx context.Context, fromUserID string, toUserName string, amount int) error
}

type CoinTransferService struct {
	repo CoinTransferRepository
}

func NewCoinTransferService(repo CoinTransferRepository) *CoinTransferService {
	return &CoinTransferService{repo: repo}
}

func (s *CoinTransferService) SendCoins(ctx context.Context, fromUserID string, toUserName string, amount int) error {
	return s.repo.SendCoins(ctx, fromUserID, toUserName, amount)
}
