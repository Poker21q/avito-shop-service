package service

import (
	"context"
	"errors"
	"merch/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserInfo(ctx context.Context, userID string) (*domain.UserInfo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*domain.UserInfo), args.Error(1)
}

func TestUserService_GetUserInfo(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockReturn     *domain.UserInfo
		mockError      error
		expectedError  error
		expectedResult *domain.UserInfo
	}{
		{
			name:   "success",
			userID: "123",
			mockReturn: &domain.UserInfo{
				CoinBalance: 500,
				Inventory: []domain.UserInventory{
					{UserID: "user123", MerchName: "t-shirt", Quantity: 2},
					{UserID: "user123", MerchName: "mug", Quantity: 5},
				},
				CoinHistoryReceived: []domain.CoinTransfer{
					{FromUserID: "user456", Amount: 100, TransactionType: "received"},
				},
				CoinHistorySent: []domain.CoinTransfer{
					{ToUserID: "user789", Amount: 50, TransactionType: "sent"},
				},
			},
			mockError:     nil,
			expectedError: nil,
			expectedResult: &domain.UserInfo{
				CoinBalance: 500,
				Inventory: []domain.UserInventory{
					{UserID: "user123", MerchName: "t-shirt", Quantity: 2},
					{UserID: "user123", MerchName: "mug", Quantity: 5},
				},
				CoinHistoryReceived: []domain.CoinTransfer{
					{FromUserID: "user456", Amount: 100, TransactionType: "received"},
				},
				CoinHistorySent: []domain.CoinTransfer{
					{ToUserID: "user789", Amount: 50, TransactionType: "sent"},
				},
			},
		},
		{
			name:           "user not found",
			userID:         "456",
			mockReturn:     nil,
			mockError:      domain.ErrNotFound,
			expectedError:  domain.ErrNotFound,
			expectedResult: nil,
		},
		{
			name:           "internal server error",
			userID:         "789",
			mockReturn:     nil,
			mockError:      domain.ErrInternalServerError,
			expectedError:  domain.ErrInternalServerError,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockRepo.On("GetUserInfo", mock.Anything, tt.userID).Return(tt.mockReturn, tt.mockError)

			service := NewUserService(mockRepo)

			result, err := service.GetUserInfo(context.Background(), tt.userID)

			assert.Equal(t, tt.expectedResult, result)
			assert.True(t, errors.Is(err, tt.expectedError))
			mockRepo.AssertExpectations(t)
		})
	}
}
