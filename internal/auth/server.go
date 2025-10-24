package auth

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func Start(port, jwtSecret, ttlStr string) {
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		ttl = 24 * time.Hour
	}

	repo := NewMemRepo()
	svc := NewService(repo, jwtSecret, ttl)

	r := mux.NewRouter()
	authRouter := r.PathPrefix("/auth").Subrouter()
	NewHandler(svc).Register(authRouter)

	log.Printf("[auth] :%s (ttl=%s)", port, ttl.String())
	log.Fatal(http.ListenAndServe(":"+port, r))
}
