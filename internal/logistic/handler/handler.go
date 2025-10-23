package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"a7/internal/logistic/repo"
	"a7/internal/logistic/service"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router *chi.Mux, repo *repo.Repo, svc *service.Service) {
	router.Get("/logistic/points", func(w http.ResponseWriter, r *http.Request) {
		list, err := repo.ListLogisticPoints()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(list)
	})
	router.Post("/logistic/points", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Title, Address string
			Lat, Lon       *float64
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		id, err := repo.CreateLogisticPoint(strings.TrimSpace(req.Title), strings.TrimSpace(req.Address), req.Lat, req.Lon)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": id})
	})

	router.Post("/logistic/shipments", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			MaxWeightKg, MaxVolumeM3 float64
			DepartureAt              time.Time
			Route                    []struct {
				PointID int64
				Ordinal int
			}
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if len(req.Route) != 2 {
			http.Error(w, "ровно 2 маршрутные точки", 400)
			return
		}
		var p [2]int64
		for _, rp := range req.Route {
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
		id, err := repo.CreateShipment(req.MaxWeightKg, req.MaxVolumeM3, req.DepartureAt)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, _, la1, lo1, err := repo.GetPoint(p[0])
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_, _, la2, lo2, err := repo.GetPoint(p[1])
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_ = repo.InsertRoutePoint(id, p[0], req.DepartureAt, 1)
		arrive2 := svc.PlannedArrive(req.DepartureAt, la1, lo1, la2, lo2)
		_ = repo.InsertRoutePoint(id, p[1], arrive2, 2)
		json.NewEncoder(w).Encode(map[string]any{"id": id, "status": "PLANNED"})
	})

	router.Get("/logistic/shipments", func(w http.ResponseWriter, r *http.Request) {
		list, err := repo.ListShipments()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(list)
	})

	router.Get("/logistic/shipments/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		m, err := repo.GetShipment(id)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		json.NewEncoder(w).Encode(m)
	})

	router.Post("/logistic/shipments/{id}/send", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err := repo.MarkShipmentDeparted(id); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		rows, err := repo.DB.Query(`SELECT application_external_id FROM assignments WHERE shipment_id=$1`, id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		client := &http.Client{Timeout: 5 * time.Second}
		for rows.Next() {
			var aid int64
			if err := rows.Scan(&aid); err == nil {
				url := fmt.Sprintf("%s/internal/applications/%d/mark_in_transit", svc.OfficeURL, aid)
				resp, err := client.Post(url, "application/json", nil)
				if err == nil {
					_, _ = io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
			}
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "DEPARTED"})
	})

	router.Get("/logistic/assignments", func(w http.ResponseWriter, r *http.Request) {
		rows, err := repo.DB.Query(`SELECT id,shipment_id,application_external_id,weight_kg,volume_m3,created_at FROM assignments ORDER BY id DESC`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		out := []map[string]any{}
		for rows.Next() {
			var id, sid, aid int64
			var w, v float64
			var created time.Time
			if err := rows.Scan(&id, &sid, &aid, &w, &v, &created); err == nil {
				out = append(out, map[string]any{"id": id, "shipment_id": sid, "application_external_id": aid, "weight_kg": w, "volume_m3": v, "created_at": created})
			}
		}
		json.NewEncoder(w).Encode(out)
	})

	router.Get("/logistic/status/applications", func(w http.ResponseWriter, r *http.Request) {
		ids := r.URL.Query().Get("ids")
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
