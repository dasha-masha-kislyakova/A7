package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"a7/internal/logistic"
	"a7/internal/logistic/repo"
	"a7/internal/logistic/service"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router *chi.Mux, store *repo.Repo, svc *service.Service) {
	// --- Logistic points

	router.Get("/logistic/points", func(w http.ResponseWriter, req *http.Request) {
		list, err := store.ListLogisticPoints()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(list) // []logistic.LogisticPoint
	})

	router.Post("/logistic/points", func(w http.ResponseWriter, req *http.Request) {
		var r struct {
			Title   string   `json:"title"`
			Address string   `json:"address"`
			Lat     *float64 `json:"lat"`
			Lon     *float64 `json:"lon"`
		}
		if err := json.NewDecoder(req.Body).Decode(&r); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		id, err := store.CreateLogisticPoint(strings.TrimSpace(r.Title), strings.TrimSpace(r.Address), r.Lat, r.Lon)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": id})
	})

	// --- Shipments

	router.Post("/logistic/shipments", func(w http.ResponseWriter, req *http.Request) {
		var r struct {
			MaxWeightKg float64 `json:"max_weight_kg"`
			MaxVolumeM3 float64 `json:"max_volume_m3"`
			DepartureAt time.Time
			Route       []struct {
				PointID int64 `json:"point_id"`
				Ordinal int   `json:"ordinal"`
			} `json:"route"`
		}
		if err := json.NewDecoder(req.Body).Decode(&r); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if len(r.Route) != 2 {
			http.Error(w, "ровно 2 маршрутные точки", 400)
			return
		}
		var p [2]int64
		for _, rp := range r.Route {
			if rp.Ordinal < 1 || rp.Ordinal > 2 {
				http.Error(w, "ordinal must be 1 or 2", 400)
				return
			}
			p[rp.Ordinal-1] = rp.PointID
		}
		if p[0] == 0 || p[1] == 0 || p[0] == p[1] {
			http.Error(w, "нужны 2 разные точки", 400)
			return
		}

		id, err := store.CreateShipment(r.MaxWeightKg, r.MaxVolumeM3, r.DepartureAt)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		// route points
		p1, err := store.GetPoint(p[0])
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		p2, err := store.GetPoint(p[1])
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_ = store.InsertRoutePoint(id, p[0], r.DepartureAt, 1)
		arrive2 := svc.PlannedArrive(r.DepartureAt, p1.Lat, p1.Lon, p2.Lat, p2.Lon)
		_ = store.InsertRoutePoint(id, p[1], arrive2, 2)

		json.NewEncoder(w).Encode(map[string]any{"id": id, "status": "PLANNED"})
	})

	router.Get("/logistic/shipments", func(w http.ResponseWriter, req *http.Request) {
		list, err := store.ListShipments() // []logistic.Shipment
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(list)
	})

	router.Get("/logistic/shipments/{id}", func(w http.ResponseWriter, req *http.Request) {
		id, _ := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		sh, route, err := store.GetShipment(id) // logistic.Shipment, []logistic.RoutePoint
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		type resp struct {
			logistic.Shipment
			Route []logistic.RoutePoint `json:"route"`
		}
		json.NewEncoder(w).Encode(resp{Shipment: sh, Route: route})
	})

	router.Post("/logistic/shipments/{id}/send", func(w http.ResponseWriter, req *http.Request) {
		id, _ := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		if err := store.MarkShipmentDeparted(id); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		ass, err := store.ListAssignmentsByShipment(id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		// notify office to mark IN_TRANSIT
		client := &http.Client{Timeout: 5 * time.Second}
		for _, a := range ass {
			url := fmt.Sprintf("%s/internal/applications/%d/mark_in_transit", svc.OfficeURL, a.ApplicationExternalID)
			resp, err := client.Post(url, "application/json", nil)
			if err == nil {
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "DEPARTED"})
	})

	// --- Assignments

	router.Get("/logistic/assignments", func(w http.ResponseWriter, req *http.Request) {
		list, err := store.ListAssignments() // []logistic.Assignment
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(list)
	})

	// --- Status passthrough (для логистической точки)

	router.Get("/logistic/status/applications", func(w http.ResponseWriter, req *http.Request) {
		ids := req.URL.Query().Get("ids")
		if ids == "" {
			json.NewEncoder(w).Encode([]any{})
			return
		}
		url := fmt.Sprintf("%s/internal/applications/status?ids=%s", svc.OfficeURL, ids)
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})
}
