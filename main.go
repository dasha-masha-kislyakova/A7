package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ah "a7/internal/auth/handler"
	authrepo "a7/internal/auth/repo"
	authsvc "a7/internal/auth/service"
	"a7/internal/common"
	loghandler "a7/internal/logistic/handler"
	logrepo "a7/internal/logistic/repo"
	logsvc "a7/internal/logistic/service"
	offhandler "a7/internal/office/handler"
	offrepo "a7/internal/office/repo"
	offsvc "a7/internal/office/service"
	"a7/internal/proxy"

	"github.com/go-chi/chi/v5"
)

func main() {
	service := getEnv("SERVICE", "proxy")
	switch service {
	case "auth":
		runAuth()
	case "office":
		runOffice()
	case "logistic":
		runLogistic()
	case "proxy":
		runProxy()
	case "all":
		runAll()
	default:
		log.Fatalf("unknown SERVICE=%s", service)
	}
}

func runAuth() {
	port := getEnv("PORT", "8083")
	secret := mustEnv("JWT_SECRET")
	repo := authrepo.NewInMemoryUsers()
	svc := authsvc.New(repo, secret)
	mux := http.NewServeMux()
	ah.RegisterRoutes(mux, svc)
	log.Printf("auth on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func runOffice() {
	port := getEnv("PORT", "8081")
	dsn := mustEnv("DB_DSN")
	db := common.MustConnectPostgres(dsn)
	if err := common.RunMigrations(db, "./migrations/office"); err != nil {
		log.Fatal(err)
	}
	repo := offrepo.New(db)
	svc := offsvc.New(repo)
	r := chi.NewRouter()
	offhandler.RegisterRoutes(r, svc)
	log.Printf("office on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func runLogistic() {
	port := getEnv("PORT", "8082")
	dsn := mustEnv("DB_DSN")
	internalOffice := getEnv("OFFICE_INTERNAL_URL", "http://office:8081")
	interval := atoi(getEnv("PLANNER_INTERVAL", "15"))
	db := common.MustConnectPostgres(dsn)
	if err := common.RunMigrations(db, "./migrations/logistic"); err != nil {
		log.Fatal(err)
	}
	repo := logrepo.New(db)
	svc := logsvc.New(repo, internalOffice, interval)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go svc.StartPlanner(ctx)
	r := chi.NewRouter()
	loghandler.RegisterRoutes(r, repo, svc)
	log.Printf("logistic on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func runProxy() {
	port := getEnv("PORT", "8080")
	authURL := getEnv("AUTH_URL", "http://auth:8083")
	officeURL := getEnv("OFFICE_URL", "http://office:8081")
	logisticURL := getEnv("LOGISTIC_URL", "http://logistic:8082")
	secret := mustEnv("JWT_SECRET")
	feDir := getEnv("FE_DIR", "./FE")
	log.Fatal(proxy.Start(":"+port, authURL, officeURL, logisticURL, secret, feDir))
}

func runAll() {

	secret := getEnv("JWT_SECRET", "dev-secret")
	feDir := getEnv("FE_DIR", "./FE")

	authMux := http.NewServeMux()
	authSvc := authsvc.New(authrepo.NewInMemoryUsers(), secret)
	ah.RegisterRoutes(authMux, authSvc)
	go func() { log.Fatal(http.ListenAndServe(":8083", authMux)) }()

	offdb := common.MustConnectPostgres(getEnv("OFFICE_DSN", "postgres://postgres:postgres@localhost:5433/office?sslmode=disable"))
	if err := common.RunMigrations(offdb, "./migrations/office"); err != nil {
		log.Fatal(err)
	}
	offR := offrepo.New(offdb)
	offS := offsvc.New(offR)
	offMux := chi.NewRouter()
	offhandler.RegisterRoutes(offMux, offS)
	go func() { log.Fatal(http.ListenAndServe(":8081", offMux)) }()

	logdb := common.MustConnectPostgres(getEnv("LOGISTIC_DSN", "postgres://postgres:postgres@localhost:5434/logistic?sslmode=disable"))
	if err := common.RunMigrations(logdb, "./migrations/logistic"); err != nil {
		log.Fatal(err)
	}
	logR := logrepo.New(logdb)
	logS := logsvc.New(logR, "http://localhost:8081", atoi(getEnv("PLANNER_INTERVAL", "15")))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go logS.StartPlanner(ctx)
	logMux := chi.NewRouter()
	loghandler.RegisterRoutes(logMux, logR, logS)
	go func() { log.Fatal(http.ListenAndServe(":8082", logMux)) }()

	go func() {
		log.Fatal(proxy.Start(":8080", "http://localhost:8083", "http://localhost:8081", "http://localhost:8082", secret, feDir))
	}()

	log.Print("all services started: proxy:8080, auth:8083, office:8081, logistic:8082")
	// wait for signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	time.Sleep(200 * time.Millisecond)
}

func atoi(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i] - '0'
		if c <= 9 {
			n = n*10 + int(c)
		}
	}
	return n
}
func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing env %s", k)
	}
	return v
}
