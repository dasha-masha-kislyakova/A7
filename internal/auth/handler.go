package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Handler struct{ svc Service }

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r *mux.Router) {
	// r — суброутер с префиксом /auth
	r.HandleFunc("/login", h.login).Methods("POST")
	r.HandleFunc("/verify", h.verify).Methods("GET") // простая проверка, что токен принят прокси/клиентом
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResp struct {
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
	ExpiresIn   time.Duration `json:"expires_in"`
	Role        string        `json:"role"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	token, role, ttl, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	respondJSON(w, http.StatusOK, loginResp{
		AccessToken: token,
		TokenType:   "bearer",
		ExpiresIn:   ttl,
		Role:        role,
	})
}

func (h *Handler) verify(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func respondJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
