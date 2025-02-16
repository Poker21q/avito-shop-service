package service

import (
	"context"
	"merch/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCoinTransferRepository struct {
	mock.Mock
}

func (m *MockCoinTransferRepository) SendCoins(ctx context.Context, fromUserID string, toUserName string, amount int) error {
	args := m.Called(ctx, fromUserID, toUserName, amount)
	return args.Error(0)
}

func TestCoinTransferService_SendCoins(t *testing.T) {
	tests := []struct {
		name          string
		fromUserID    string
		toUserName    string
		amount        int
		mockError     error
		expectedError error
	}{
		{
			name:          "success",
			fromUserID:    "123",
			toUserName:    "user456",
			amount:        100,
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "insufficient funds",
			fromUserID:    "123",
			toUserName:    "user456",
			amount:        1000,
			mockError:     domain.ErrInsufficientFunds,
			expectedError: domain.ErrInsufficientFunds,
		},
		{
			name:          "user not found",
			fromUserID:    "123",
			toUserName:    "user999",
			amount:        100,
			mockError:     domain.ErrNotFound,
			expectedError: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCoinTransferRepository)
			mockRepo.On("SendCoins", mock.Anything, tt.fromUserID, tt.toUserName, tt.amount).Return(tt.mockError)

			service := NewCoinTransferService(mockRepo)

			err := service.SendCoins(context.Background(), tt.fromUserID, tt.toUserName, tt.amount)

			assert.Equal(t, tt.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}
