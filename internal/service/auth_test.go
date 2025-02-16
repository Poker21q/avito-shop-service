package service

import (
	"context"
	"errors"
	"merch/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthRepository struct {
	mock.Mock
}

func (m *MockAuthRepository) IsUserExists(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthRepository) CreateUser(ctx context.Context, username, passwordHash string) (string, error) {
	args := m.Called(ctx, username, passwordHash)
	return args.String(0), args.Error(1)
}

func (m *MockAuthRepository) Auth(ctx context.Context, username, passwordHash string) (string, error) {
	args := m.Called(ctx, username, passwordHash)
	return args.String(0), args.Error(1)
}

func TestAuthService_Auth(t *testing.T) {
	tests := []struct {
		name                string
		username            string
		password            string
		mockIsUserExists    bool
		mockCreateUserID    string
		mockCreateUserError error
		mockAuthUserID      string
		mockAuthError       error
		expectedError       error
	}{
		{
			name:                "register new user",
			username:            "user1",
			password:            "password1",
			mockIsUserExists:    false,
			mockCreateUserID:    "newUserID",
			mockCreateUserError: nil,
			mockAuthUserID:      "",
			mockAuthError:       nil,
			expectedError:       nil,
		},
		{
			name:                "existing user, valid credentials",
			username:            "user2",
			password:            "password2",
			mockIsUserExists:    true,
			mockCreateUserID:    "",
			mockCreateUserError: nil,
			mockAuthUserID:      "existingUserID",
			mockAuthError:       nil,
			expectedError:       nil,
		},
		{
			name:                "user not found, registration failed",
			username:            "user3",
			password:            "password3",
			mockIsUserExists:    false,
			mockCreateUserID:    "",
			mockCreateUserError: errors.New("user creation error"),
			mockAuthUserID:      "",
			mockAuthError:       nil,
			expectedError:       domain.ErrInternalServerError,
		},
		{
			name:                "invalid credentials",
			username:            "user4",
			password:            "wrongpassword",
			mockIsUserExists:    true,
			mockCreateUserID:    "",
			mockCreateUserError: nil,
			mockAuthUserID:      "",
			mockAuthError:       errors.New("invalid credentials"),
			expectedError:       domain.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAuthRepository)

			mockRepo.On("IsUserExists", mock.Anything, tt.username).Return(tt.mockIsUserExists, nil)

			if tt.mockIsUserExists == false {
				mockRepo.On("CreateUser", mock.Anything, tt.username, hashPassword(tt.password)).Return(tt.mockCreateUserID, tt.mockCreateUserError)
			}

			if tt.mockIsUserExists == true {
				mockRepo.On("Auth", mock.Anything, tt.username, hashPassword(tt.password)).Return(tt.mockAuthUserID, tt.mockAuthError)
			}

			service := NewAuthService(mockRepo, "secret")

			_, err := service.Auth(context.Background(), tt.username, tt.password)

			assert.True(t, errors.Is(err, tt.expectedError))
			mockRepo.AssertExpectations(t)
		})
	}
}
