package office

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Handler struct{ svc Service }

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r *mux.Router) {
	// здесь r — уже суброутер с префиксом /office
	r.HandleFunc("/applications", h.create).Methods("POST")
	r.HandleFunc("/applications/{id:[0-9]+}", h.getByID).Methods("GET")
	r.HandleFunc("/applications/{id:[0-9]+}/status", h.updateStatus).Methods("POST")
	r.HandleFunc("/applications", h.list).Methods("GET")
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	app, err := h.svc.Create(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	respondJSON(w, http.StatusCreated, app)
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	app, err := h.svc.Get(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	respondJSON(w, http.StatusOK, app)
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateStatus(r.Context(), id, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	var st *string
	if v := r.URL.Query().Get("status"); v != "" {
		st = &v
	}
	list, err := h.svc.List(r.Context(), st)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, list)
}

func respondJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
