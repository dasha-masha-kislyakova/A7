package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gorilla/mux"
)

func Start(port, officeURL, logisticURL, authURL, auditURL, feDir string) {
	r := mux.NewRouter()
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	r.PathPrefix("/office/").Handler(reverseProxy(officeURL))
	r.PathPrefix("/logistic/").Handler(reverseProxy(logisticURL))
	r.PathPrefix("/auth/").Handler(reverseProxy(authURL))
	r.PathPrefix("/audit/").Handler(reverseProxy(auditURL))

	// FE статика (если есть)
	if stat, err := os.Stat(feDir); err == nil && stat.IsDir() {
		r.PathPrefix("/").Handler(http.FileServer(http.Dir(feDir)))
	} else {
		r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		})
	}

	log.Printf("[proxy] :%s (office=%s, logistic=%s, auth=%s, audit=%s, FE=%s)", port, officeURL, logisticURL, authURL, auditURL, feDir)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func reverseProxy(target string) http.Handler {
	u, err := url.Parse(target)
	if err != nil {
		log.Fatalf("bad proxy target %q: %v", target, err)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Host = u.Host
		proxy.ServeHTTP(w, r)
	})
}
