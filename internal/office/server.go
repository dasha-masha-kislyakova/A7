package office

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"template/internal/db"
	"template/internal/middleware/authmw"
)

func Start(port, dsn, jwtSecret string) {
	dbc := db.MustConnect(dsn)
	repo := NewRepo(dbc)
	if err := repo.EnsureSchema(context.Background()); err != nil {
		log.Fatalf("office ensure schema: %v", err)
	}
	svc, _ := NewService(repo)

	r := mux.NewRouter()
	// подпрефикс /office
	officeRouter := r.PathPrefix("/office").Subrouter()

	// доступ только менеджеру офиса
	mw := authmw.New(jwtSecret)
	officeRouter.Use(mw.RequireRoles("office_manager"))

	NewHandler(svc).Register(officeRouter)

	log.Printf("[office] :%s (dsn=%s)", port, dsn)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
