package service

import (
	"a7/internal/logistic/repo"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

type Service struct {
	R         *repo.Repo
	OfficeURL string
	interval  time.Duration
}

func New(r *repo.Repo, officeURL string, intervalSec int) *Service {
	return &Service{R: r, OfficeURL: officeURL, interval: time.Duration(intervalSec) * time.Second}
}

func (s *Service) StartPlanner(ctx context.Context) {
	if s.interval <= 0 {
		return
	}
	t := time.NewTicker(s.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			_ = s.tick()
		}
	}
}

func (s *Service) tick() error {
	shipments, err := s.R.GetOpenShipments()
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	for _, sh := range shipments {
		sid := sh["id"].(int64)
		dep := sh["departure_at"].(time.Time)
		p1 := sh["p1"].(int64)
		p2 := sh["p2"].(int64)
		url := fmt.Sprintf("%s/internal/applications?pickup_point_id=%d&dropoff_point_id=%d&before=%s", s.OfficeURL, p1, p2, dep.Format(time.RFC3339))
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("office: %d: %s", resp.StatusCode, string(b))
		}
		var apps []map[string]any
		if err := json.Unmarshal(b, &apps); err != nil {
			return err
		}
		curW, curV, _ := s.R.CurrentLoad(sid)
		maxW := sh["max_weight_kg"].(float64)
		maxV := sh["max_volume_m3"].(float64)
		for _, a := range apps {
			aw := a["total_weight"].(float64)
			av := a["total_volume"].(float64)
			if curW+aw <= maxW && curV+av <= maxV {
				_ = s.R.AddAssignment(sid, int64(a["id"].(float64)), aw, av, p1, p2)
				curW += aw
				curV += av
			}
		}
	}
	return nil
}

func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	la1 := lat1 * math.Pi / 180
	la2 := lat2 * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(la1)*math.Cos(la2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
func (s *Service) PlannedArrive(departure time.Time, la1, lo1, la2, lo2 *float64) time.Time {
	if la1 != nil && lo1 != nil && la2 != nil && lo2 != nil {
		d := haversineKm(*la1, *lo1, *la2, *lo2)
		h := d / 60.0
		return departure.Add(time.Duration(h * float64(time.Hour)))
	}
	return departure.Add(2 * time.Hour)
}
