package app

import (
	"log"
	"os"

	"template/internal/audit"
	"template/internal/auth"
	"template/internal/logistic"
	"template/internal/office"
	"template/internal/proxy"
)

type App struct{}

func New() *App { return &App{} }

func (a *App) Run() {
	switch getenv("SERVICE", "proxy") {
	case "proxy":
		proxy.Start(
			getenv("PORT", "8080"),
			getenv("OFFICE_URL", "http://office:8081"),
			getenv("LOGISTIC_URL", "http://logistic:8082"),
			getenv("AUTH_URL", "http://auth:8083"),
			getenv("AUDIT_URL", "http://audit:8084"),
			getenv("FE_DIR", "/app/FE"),
		)
	case "office":
		office.Start(
			getenv("PORT", "8081"),
			getenv("OFFICE_DSN", "postgres://postgres:postgres@office-db:5432/office?sslmode=disable"),
			getenv("JWT_SECRET", "dev-secret"),
		)
	case "logistic":
		logistic.Start(
			getenv("PORT", "8082"),
			getenv("LOGISTIC_DSN", "postgres://postgres:postgres@logistic-db:5432/logistic?sslmode=disable"),
			getenv("OFFICE_INTERNAL_URL", "http://office:8081"),
			getenv("JWT_SECRET", "dev-secret"),
		)
	case "auth":
		auth.Start(
			getenv("PORT", "8083"),
			getenv("JWT_SECRET", "dev-secret"),
			getenv("TOKEN_TTL", "24h"),
		)
	case "audit":
		audit.Start(
			getenv("PORT", "8084"),
		)
	default:
		log.Fatalf("unknown SERVICE")
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
