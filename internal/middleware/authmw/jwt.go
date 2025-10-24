package authmw

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type MW struct {
	secret []byte
}

func New(secret string) *MW { return &MW{secret: []byte(secret)} }

type Claims struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (m *MW) parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if c, ok := token.Claims.(*Claims); ok && token.Valid {
		return c, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return ""
	}
	return strings.TrimSpace(h[len("Bearer "):])
}

// RequireToken — просто валидный токен
func (m *MW) RequireToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok := bearer(r)
		if tok == "" {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		if _, err := m.parse(tok); err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRoles — проверка на конкретные роли
func (m *MW) RequireRoles(roles ...string) mux.MiddlewareFunc {
	allowed := map[string]struct{}{}
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tok := bearer(r)
			if tok == "" {
				http.Error(w, "missing bearer token", http.StatusUnauthorized)
				return
			}
			c, err := m.parse(tok)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			if len(allowed) > 0 {
				if _, ok := allowed[c.Role]; !ok {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
