package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// AuthMiddleware - middleware для проверки токена
type AuthMiddleware struct {
	authServiceURL string
}

func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Authorization token required", http.StatusUnauthorized)
			return
		}

		// Валидация токена через микросервис авторизации
		managerID, err := m.validateToken(r.Context(), token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Добавляем ID менеджера в контекст
		ctx := context.WithValue(r.Context(), "manager_id", managerID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) validateToken(ctx context.Context, token string) (int64, string, error) {
	// Реализация проверки токена через HTTP запрос к сервису авторизации
	// Возвращает ID менеджера и его роль
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", m.authServiceURL+"/validate", nil)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Authorization", token)

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("invalid token")
	}

	var result struct {
		UserID int64  `json:"user_id"`
		Role   string `json:"role"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, "", err
	}

	return result.UserID, result.Role, nil
}

// AuditMiddleware - middleware для автоматического логирования HTTP запросов
type AuditMiddleware struct {
	auditService AuditService
	serviceName  string
}

func NewAuditMiddleware(auditService AuditService, serviceName string) *AuditMiddleware {
	return &AuditMiddleware{
		auditService: auditService,
		serviceName:  serviceName,
	}
}

func (m *AuditMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Перехватываем response body
		rw := &responseWriter{ResponseWriter: w, body: bytes.NewBufferString("")}

		// Выполняем запрос
		next.ServeHTTP(rw, r)

		// Логируем событие
		go m.logAuditEvent(r, rw, start)
	})
}

func (m *AuditMiddleware) logAuditEvent(r *http.Request, rw *responseWriter, start time.Time) {
	processingTime := time.Since(start).Milliseconds()

	// Получаем UserID из контекста, если есть
	var userID *int64
	if ctxUserID, ok := r.Context().Value("user_id").(int64); ok {
		userID = &ctxUserID
	}

	event := &AuditEventRequest{
		RequestService:   m.getRequestService(r),
		ResponseService:  m.serviceName,
		URI:              r.URL.String(),
		HTTPStatus:       rw.status,
		EventDate:        start,
		ProcessingTimeMs: processingTime,
		UserID:           userID,
		RequestBody:      m.getRequestBody(r),
		ResponseBody:     rw.body.String(),
	}

	if err := m.auditService.LogEvent(context.Background(), event); err != nil {
		log.Printf("Failed to log audit event: %v", err)
	}
}

func (m *AuditMiddleware) getRequestService(r *http.Request) string {
	// В реальной реализации можно извлекать из заголовков
	return r.Header.Get("X-Service-Name")
}

func (m *AuditMiddleware) getRequestBody(r *http.Request) string {
	if r.Body == nil {
		return ""
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return ""
	}

	// Восстанавливаем body для дальнейшего использования
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return string(body)
}

// responseWriter - перехватчик response
type responseWriter struct {
	http.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}
