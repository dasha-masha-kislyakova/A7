package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	Port        string
	AuthURL     string
	OfficeURL   string
	LogisticURL string
	FEDir       string
	JWTSecret   string
}

// Start — универсальный запуск прокси.
// Поддерживает вызовы с разным числом аргументов (как у вас в main.go).
// Если какие-то аргументы не переданы — берутся из переменных окружения.
func Start(args ...string) error {
	// defaults из окружения
	cfg := Config{
		Port:        getEnv("PORT", "8080"),
		FEDir:       getEnv("FE_DIR", "./FE"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		AuthURL:     getEnv("AUTH_URL", "http://localhost:8083"),
		OfficeURL:   getEnv("OFFICE_URL", "http://localhost:8081"),
		LogisticURL: getEnv("LOGISTIC_URL", "http://localhost:8082"),
	}

	// если main.go передаёт параметры позиционно — примем их
	switch len(args) {
	case 6:
		cfg.Port, cfg.FEDir, cfg.JWTSecret, cfg.AuthURL, cfg.OfficeURL, cfg.LogisticURL = args[0], args[1], args[2], args[3], args[4], args[5]
	case 5:
		cfg.Port, cfg.FEDir, cfg.JWTSecret, cfg.AuthURL, cfg.OfficeURL = args[0], args[1], args[2], args[3], args[4]
	case 4:
		cfg.Port, cfg.FEDir, cfg.JWTSecret, cfg.AuthURL = args[0], args[1], args[2], args[3]
	case 3:
		cfg.Port, cfg.FEDir, cfg.JWTSecret = args[0], args[1], args[2]
	case 2:
		cfg.Port, cfg.FEDir = args[0], args[1]
	case 1:
		cfg.Port = args[0]
	}

	mux := newMux(cfg)
	log.Printf("proxy listening on :%s (auth=%s office=%s logistic=%s fe=%s)",
		cfg.Port, cfg.AuthURL, cfg.OfficeURL, cfg.LogisticURL, cfg.FEDir)
	return http.ListenAndServe(":"+cfg.Port, mux)
}

func newMux(cfg Config) *http.ServeMux {
	mux := http.NewServeMux()

	authURL := mustParse(cfg.AuthURL)
	officeURL := mustParse(cfg.OfficeURL)
	logisticURL := mustParse(cfg.LogisticURL)

	// /auth/* — без JWT
	mux.Handle("/auth/", httputil.NewSingleHostReverseProxy(authURL))

	// /office/* и /logistic/* — с проверкой JWT
	mux.Handle("/office/", jwtMiddleware(httputil.NewSingleHostReverseProxy(officeURL), cfg.JWTSecret))
	mux.Handle("/logistic/", jwtMiddleware(httputil.NewSingleHostReverseProxy(logisticURL), cfg.JWTSecret))

	// статика FE
	fs := http.FileServer(http.Dir(cfg.FEDir))
	mux.Handle("/", fs)

	// health
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

func jwtMiddleware(next http.Handler, secret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if secret == "" {
			http.Error(w, "JWT secret is empty", http.StatusUnauthorized)
			return
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "missing Bearer token", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer"))
		_, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			// допускаем только HS256
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func mustParse(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		log.Fatalf("bad url %q: %v", raw, err)
	}
	return u
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
