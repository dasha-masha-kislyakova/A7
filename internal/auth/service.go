package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service interface {
	Login(username, password string) (string, string, time.Duration, error)
}

type service struct {
	repo   Repo
	secret []byte
	ttl    time.Duration
}

func NewService(repo Repo, jwtSecret string, ttl time.Duration) Service {
	return &service{repo: repo, secret: []byte(jwtSecret), ttl: ttl}
}

func (s *service) Login(username, password string) (string, string, time.Duration, error) {
	u, ok := s.repo.ByUsername(username)
	if !ok || u.Password != password {
		return "", "", 0, errors.New("invalid credentials")
	}
	claims := jwt.MapClaims{
		"sub":  u.Username,
		"role": u.Role,
		"exp":  time.Now().Add(s.ttl).Unix(),
		"iss":  "auth",
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := t.SignedString(s.secret)
	if err != nil {
		return "", "", 0, err
	}
	return tokenStr, u.Role, s.ttl, nil
}
