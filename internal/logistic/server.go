package logistic

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"template/internal/db"
	"template/internal/middleware/authmw"
)

func Start(port, dsn, officeInternalURL, jwtSecret string) {
	dbc := db.MustConnect(dsn)
	repo := NewRepo(dbc)
	if err := repo.EnsureSchema(context.Background()); err != nil {
		log.Fatalf("logistic ensure schema: %v", err)
	}
	svc, _ := NewService(repo, officeInternalURL)

	r := mux.NewRouter()
	logRouter := r.PathPrefix("/logistic").Subrouter()

	// доступ только менеджеру логистической точки
	mw := authmw.New(jwtSecret)
	logRouter.Use(mw.RequireRoles("logistics_manager"))

	NewHandler(svc).Register(logRouter)

	log.Printf("[logistic] :%s (dsn=%s, officeCB=%s)", port, dsn, officeInternalURL)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
