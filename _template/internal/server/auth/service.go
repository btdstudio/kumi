package auth

import (
	"time"

	"goph/internal/infrastructure"
)

const resetTokenTTL = 1 * time.Hour

type Service struct {
	SecretKey string
}

func NewService(secretKey string) *Service {
	return &Service{SecretKey: secretKey}
}

func (s *Service) CreateResetToken(userID string) (string, error) {
	return infrastructure.CreateToken(userID, s.SecretKey, resetTokenTTL)
}

func (s *Service) VerifyResetToken(token string) (string, error) {
	return infrastructure.VerifyToken(token, s.SecretKey)
}
