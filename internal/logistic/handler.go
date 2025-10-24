package logistic

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Handler struct{ svc Service }

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(r *mux.Router) {
	// здесь r — уже суброутер с префиксом /logistic
	r.HandleFunc("/applications", h.listApps).Methods("GET")
	r.HandleFunc("/applications/{id:[0-9]+}", h.getApp).Methods("GET")
	r.HandleFunc("/applications/{id:[0-9]+}/status", h.updateAppStatus).Methods("POST")

	r.HandleFunc("/routes", h.createRoute).Methods("POST")
	r.HandleFunc("/routes/{routeId:[0-9]+}/assign/{applicationId:[0-9]+}", h.assign).Methods("POST")
	r.HandleFunc("/routes/{routeId:[0-9]+}/send", h.sendRoute).Methods("POST")

	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
}

func (h *Handler) listApps(w http.ResponseWriter, r *http.Request) {
	var st *string
	if v := r.URL.Query().Get("status"); v != "" {
		st = &v
	}
	out, err := h.svc.ListLogApps(r.Context(), st)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	respondJSON(w, http.StatusOK, out)
}

func (h *Handler) getApp(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	app, err := h.svc.GetLogApp(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	respondJSON(w, http.StatusOK, app)
}

func (h *Handler) updateAppStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := h.svc.UpdateLogAppStatus(r.Context(), id, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) createRoute(w http.ResponseWriter, r *http.Request) {
	var req CreateRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	route, err := h.svc.CreateRoute(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	respondJSON(w, http.StatusCreated, route)
}

func (h *Handler) assign(w http.ResponseWriter, r *http.Request) {
	routeID, _ := strconv.ParseInt(mux.Vars(r)["routeId"], 10, 64)
	appID, _ := strconv.ParseInt(mux.Vars(r)["applicationId"], 10, 64) // office app id
	if err := h.svc.AssignApp(r.Context(), routeID, appID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) sendRoute(w http.ResponseWriter, r *http.Request) {
	routeID, _ := strconv.ParseInt(mux.Vars(r)["routeId"], 10, 64)
	if err := h.svc.SendRoute(r.Context(), routeID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func respondJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
