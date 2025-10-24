package audit

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func Start(port string) {
	repo := NewMemRepo()

	r := mux.NewRouter()
	auditRouter := r.PathPrefix("/audit").Subrouter()
	NewHandler(repo).Register(auditRouter)

	log.Printf("[audit] :%s (in-memory)", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
