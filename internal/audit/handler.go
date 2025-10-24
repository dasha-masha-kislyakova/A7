package audit

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Handler struct{ repo Repo }

func NewHandler(repo Repo) *Handler { return &Handler{repo: repo} }

func (h *Handler) Register(r *mux.Router) {
	// r — суброутер с префиксом /audit
	r.HandleFunc("/events", h.create).Methods("POST")
	r.HandleFunc("/events", h.list).Methods("GET")
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var e Event
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	ev, _ := h.repo.Insert(e)
	respondJSON(w, http.StatusCreated, ev)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := Query{
		SourceService: r.URL.Query().Get("source"),
		TargetService: r.URL.Query().Get("target"),
		URI:           r.URL.Query().Get("uri"),
		UserID:        r.URL.Query().Get("user"),
		SortBy:        r.URL.Query().Get("sort"),
		SortOrder:     r.URL.Query().Get("order"),
	}
	if s := r.URL.Query().Get("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			q.HTTPStatus = &v
		}
	}
	if df := r.URL.Query().Get("from"); df != "" {
		if t, err := time.Parse(time.RFC3339, df); err == nil {
			q.DateFrom = &t
		}
	}
	if dt := r.URL.Query().Get("to"); dt != "" {
		if t, err := time.Parse(time.RFC3339, dt); err == nil {
			q.DateTo = &t
		}
	}
	if md := r.URL.Query().Get("min_dur_ms"); md != "" {
		if v, err := strconv.ParseInt(md, 10, 64); err == nil {
			q.MinDurationMs = &v
		}
	}
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			q.Page = v
		}
	}
	if ps := r.URL.Query().Get("size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil {
			q.PageSize = v
		}
	}
	items, total := h.repo.List(q)
	respondJSON(w, http.StatusOK, map[string]any{
		"total": total,
		"items": items,
	})
}

func respondJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
