package handler

import (
	"a7/internal/auth/service"
	"encoding/json"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, svc *service.Service) {
	mux.HandleFunc("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		var req struct{ Email, Password, Role string }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		token, err := svc.Register(req.Email, req.Password, req.Role)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"access_token": token})
	})
	mux.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		var req struct{ Email, Password string }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		token, err := svc.Login(req.Email, req.Password)
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"access_token": token})
	})
}
