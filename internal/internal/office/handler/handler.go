package handler

import (
	"a7/internal/office"
	"a7/internal/office/service"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func RegisterRoutes(router *chi.Mux, svc *service.Service) {
	router.Post("/office/applications", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			PickupPointID  int64                                       `json:"pickup_point_id"`
			DropoffPointID int64                                       `json:"dropoff_point_id"`
			Client         struct{ Name, INN, FIO, Tel, Email string } `json:"client"`
			Cargo          struct {
				Name                     string
				Boxes                    int64
				TotalWeight, TotalVolume float64
				SpecialRequirements      *string `json:"special_requirements"`
			} `json:"cargo"`
			Recipient struct{ NameCompany, Address, FIO, Telephone string } `json:"recipient"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		id, err := svc.CreateApplication(
			req.PickupPointID, req.DropoffPointID,
			office.Client{Name: req.Client.Name, INN: req.Client.INN, FIO: req.Client.FIO, Tel: req.Client.Tel, Email: req.Client.Email},
			office.Cargo{Name: req.Cargo.Name, Boxes: req.Cargo.Boxes, TotalWeight: req.Cargo.TotalWeight, TotalVolume: req.Cargo.TotalVolume, SpecialReq: req.Cargo.SpecialRequirements},
			office.Recipient{Name: req.Recipient.NameCompany, Address: req.Recipient.Address, FIO: req.Recipient.FIO, Telephone: req.Recipient.Telephone},
		)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": id, "status": string(office.AppNEW)})
	})

	router.Get("/office/applications", func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		list, err := svc.ListApplications(status)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(list)
	})

	router.Get("/office/applications/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		m, err := svc.GetApplication(id)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		json.NewEncoder(w).Encode(m)
	})

	router.Post("/office/applications/{id}/accept", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err := svc.Accept(id); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": string(office.AppINWORK)})
	})

	router.Post("/office/applications/{id}/deliver", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err := svc.Deliver(id); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": string(office.AppDELIVERED)})
	})

	// INTERNAL
	router.Get("/internal/applications", func(w http.ResponseWriter, r *http.Request) {
		pp, _ := strconv.ParseInt(r.URL.Query().Get("pickup_point_id"), 10, 64)
		dp, _ := strconv.ParseInt(r.URL.Query().Get("dropoff_point_id"), 10, 64)
		beforeStr := r.URL.Query().Get("before")
		var before time.Time
		if beforeStr == "" {
			before = time.Now()
		} else {
			t, err := time.Parse(time.RFC3339, beforeStr)
			if err != nil {
				http.Error(w, "bad time", 400)
				return
			} else {
				before = t
			}
		}
		list, err := svc.ListAvailableForRoute(pp, dp, before)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(list)
	})

	router.Post("/internal/applications/{id}/mark_in_transit", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err := svc.MarkInTransit(id); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"status": string(office.AppINTRANSIT)})
	})

	router.Get("/internal/applications/status", func(w http.ResponseWriter, r *http.Request) {
		idsStr := r.URL.Query().Get("ids")
		if idsStr == "" {
			json.NewEncoder(w).Encode([]any{})
			return
		}
		parts := strings.Split(idsStr, ",")
		placeholders := []string{}
		args := []any{}
		for i, p := range parts {
			placeholders = append(placeholders, "$"+strconv.Itoa(i+1))
			args = append(args, strings.TrimSpace(p))
		}
		q := "SELECT id,status FROM applications WHERE id IN (" + strings.Join(placeholders, ",") + ") ORDER BY id"
		rows, err := svc.R.DB.Query(q, args...)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		out := []map[string]any{}
		for rows.Next() {
			var id int64
			var st string
			if err := rows.Scan(&id, &st); err == nil {
				out = append(out, map[string]any{"id": id, "status": st})
			}
		}
		json.NewEncoder(w).Encode(out)
	})
}
