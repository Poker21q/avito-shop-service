package service

import (
	"context"
	"merch/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPurchaseRepository struct {
	mock.Mock
}

func (m *MockPurchaseRepository) Buy(ctx context.Context, userID, item string) error {
	args := m.Called(ctx, userID, item)
	return args.Error(0)
}

func TestPurchaseService_BuyItem(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		item          string
		mockError     error
		expectedError error
	}{
		{
			name:          "success",
			userID:        "123",
			item:          "item1",
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "purchase failed",
			userID:        "456",
			item:          "item2",
			mockError:     domain.ErrInternalServerError,
			expectedError: domain.ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockPurchaseRepository)
			mockRepo.On("Buy", mock.Anything, tt.userID, tt.item).Return(tt.mockError)

			service := NewPurchaseService(mockRepo)

			err := service.BuyItem(context.Background(), tt.userID, tt.item)

			assert.Equal(t, tt.expectedError, err)
			mockRepo.AssertExpectations(t)
		})
	}
}
