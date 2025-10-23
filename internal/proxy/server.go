package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func mustURL(env, def string) *url.URL {
	raw := os.Getenv(env)
	if raw == "" {
		raw = def
	}
	u, err := url.Parse(raw)
	if err != nil {
		log.Fatalf("bad %s=%q: %v", env, raw, err)
	}
	return u
}

func New(feDir string, jwtSecret string) *http.ServeMux {
	mux := http.NewServeMux()

	// адреса внутренних сервисов
	authURL := mustURL("AUTH_URL", "http://localhost:8083")
	officeURL := mustURL("OFFICE_URL", "http://localhost:8081")
	logisticURL := mustURL("LOGISTIC_URL", "http://localhost:8082")

	// reverse proxy
	mux.Handle("/auth/", httputil.NewSingleHostReverseProxy(authURL))
	mux.Handle("/office/", withJWT(httputil.NewSingleHostReverseProxy(officeURL), jwtSecret))
	mux.Handle("/logistic/", withJWT(httputil.NewSingleHostReverseProxy(logisticURL), jwtSecret))

	// статика FE
	fs := http.FileServer(http.Dir(feDir))
	mux.Handle("/", fs)

	// простая health-страница
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200); _, _ = w.Write([]byte("ok")) })

	return mux
}

// простая проверка наличия Authorization: Bearer <token>
// (если у вас уже есть своя валидация JWT — используйте её вместо этой заглушки)
func withJWT(next http.Handler, secret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/office/applications" && r.Method == http.MethodPost {
			// этот код — пример; в нормальном случае проверяем все защищённые ручки
		}
		if r.Header.Get("Authorization") == "" {
			http.Error(w, "missing Authorization", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
