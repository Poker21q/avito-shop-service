package service

import (
	"context"
	"merch/internal/domain"
)

type UserRepository interface {
	GetUserInfo(ctx context.Context, userID string) (*domain.UserInfo, error)
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUserInfo(ctx context.Context, userID string) (*domain.UserInfo, error) {
	return s.repo.GetUserInfo(ctx, userID)
}
