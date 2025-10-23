package repo

import (
	"a7/internal/auth"
	"errors"
	"strings"
	"sync"
)

type InMemoryUsers struct {
	mu      sync.Mutex
	next    int
	byEmail map[string]*auth.User
}

func NewInMemoryUsers() *InMemoryUsers {
	s := &InMemoryUsers{byEmail: map[string]*auth.User{}, next: 1}
	s.byEmail["office_admin@example.com"] = &auth.User{ID: s.next, Email: "office_admin@example.com", Password: "password", Role: "office_admin"}
	s.next++
	s.byEmail["logpoint_admin@example.com"] = &auth.User{ID: s.next, Email: "logpoint_admin@example.com", Password: "password", Role: "logpoint_admin"}
	s.next++
	return s
}

func (r *InMemoryUsers) Register(email, password, role string) (*auth.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	email = strings.TrimSpace(email)
	if _, ok := r.byEmail[email]; ok {
		return nil, errors.New("user exists")
	}
	if role != "office_admin" && role != "logpoint_admin" {
		return nil, errors.New("invalid role")
	}
	u := &auth.User{ID: r.next, Email: email, Password: password, Role: role}
	r.byEmail[email] = u
	r.next++
	return u, nil
}

func (r *InMemoryUsers) FindByCredentials(email, password string) (*auth.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.byEmail[strings.TrimSpace(email)]
	if !ok || u.Password != password {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}
