package service

import (
	"a7/internal/auth/repo"
	"a7/internal/common"
	"time"
)

type Service struct {
	Users     *repo.InMemoryUsers
	JWTSecret string
}

func New(users *repo.InMemoryUsers, secret string) *Service {
	return &Service{Users: users, JWTSecret: secret}
}

func (s *Service) Register(email, pass, role string) (string, error) {
	u, err := s.Users.Register(email, pass, role)
	if err != nil {
		return "", err
	}
	return common.SignJWT(s.JWTSecret, u.ID, u.Email, u.Role, 24*time.Hour)
}

func (s *Service) Login(email, pass string) (string, error) {
	u, err := s.Users.FindByCredentials(email, pass)
	if err != nil {
		return "", err
	}
	return common.SignJWT(s.JWTSecret, u.ID, u.Email, u.Role, 24*time.Hour)
}
