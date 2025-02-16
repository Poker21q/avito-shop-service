package service

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"merch/internal/domain"
	"merch/pkg/jwtutils"
	"time"
)

type AuthRepository interface {
	IsUserExists(ctx context.Context, username string) (bool, error)
	CreateUser(ctx context.Context, username, passwordHash string) (userID string, err error)
	Auth(ctx context.Context, username, passwordHash string) (userID string, err error)
}

type AuthService struct {
	repo      AuthRepository
	jwtSecret string
}

func NewAuthService(repo AuthRepository, jwtSecret string) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Auth(ctx context.Context, username, password string) (string, error) {
	passwordHash := hashPassword(password)

	exists, err := s.repo.IsUserExists(ctx, username)
	if err != nil {
		return "", domain.ErrInternalServerError
	}

	if !exists {
		userID, err := s.registerUser(ctx, username, passwordHash)
		if err != nil {
			return "", err
		}
		return s.generateJWT(userID)
	}

	userID, err := s.authenticateUser(ctx, username, passwordHash)
	if err != nil {
		return "", domain.ErrInvalidCredentials
	}

	return s.generateJWT(userID)
}

func (s *AuthService) registerUser(ctx context.Context, username, passwordHash string) (string, error) {
	userID, err := s.repo.CreateUser(ctx, username, passwordHash)
	if err != nil {
		return "", domain.ErrInternalServerError
	}
	return userID, nil
}

func (s *AuthService) authenticateUser(ctx context.Context, username, passwordHash string) (string, error) {
	userID, err := s.repo.Auth(ctx, username, passwordHash)
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (s *AuthService) generateJWT(userID string) (string, error) {
	claims := map[string]interface{}{"user_id": userID}
	token, err := jwtutils.Generate(claims, 72*time.Hour, s.jwtSecret)
	if err != nil {
		return "", domain.ErrInternalServerError
	}
	return token, nil
}

func hashPassword(password string) string {
	const salt = "da39a3ee5e6b4b"

	hash := sha1.New()
	hash.Write([]byte(password + salt))
	return hex.EncodeToString(hash.Sum(nil))
}
