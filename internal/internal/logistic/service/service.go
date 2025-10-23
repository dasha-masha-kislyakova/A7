package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"a7/internal/logistic/repo"
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
		url := fmt.Sprintf("%s/internal/applications?pickup_point_id=%d&dropoff_point_id=%d&before=%s",
			s.OfficeURL, sh.P1, sh.P2, sh.DepartureAt.Format(time.RFC3339))
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("office: %d: %s", resp.StatusCode, string(b))
		}

		var apps []struct {
			ID          int64   `json:"id"`
			TotalWeight float64 `json:"total_weight"`
			TotalVolume float64 `json:"total_volume"`
			// pickup/drop тоже есть, но нам здесь не нужны
		}
		if err := json.Unmarshal(b, &apps); err != nil {
			return err
		}

		curW, curV, _ := s.R.CurrentLoad(sh.ID)
		maxW := sh.MaxWeightKg
		maxV := sh.MaxVolumeM3

		for _, a := range apps {
			if curW+a.TotalWeight <= maxW && curV+a.TotalVolume <= maxV {
				_ = s.R.AddAssignment(sh.ID, a.ID, a.TotalWeight, a.TotalVolume, sh.P1, sh.P2)
				curW += a.TotalWeight
				curV += a.TotalVolume
			}
		}
	}
	return nil
}

func (s *Service) PlannedArrive(departure time.Time, la1, lo1, la2, lo2 *float64) time.Time {
	if la1 != nil && lo1 != nil && la2 != nil && lo2 != nil {
		d := haversineKm(*la1, *lo1, *la2, *lo2)
		h := d / 60.0
		return departure.Add(time.Duration(h * float64(time.Hour)))
	}
	return departure.Add(2 * time.Hour)
}

func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371
	dLat := (lat2 - lat1) * (3.1415926535 / 180)
	dLon := (lon2 - lon1) * (3.1415926535 / 180)
	la1 := lat1 * (3.1415926535 / 180)
	la2 := lat2 * (3.1415926535 / 180)
	a := sin2(dLat/2) + sin2(dLon/2)*cos(la1)*cos(la2)
	c := 2 * atan2sqrt(a, 1-a)
	return R * c
}

func sin2(x float64) float64    { s := fastSin(x); return s * s }
func fastSin(x float64) float64 { return x - (x*x*x)/6 }
func cos(x float64) float64     { return fastSin(1.57079632679 - x) }
func atan2sqrt(y, x float64) float64 {
	return (y - x) / (y + x + 1e-9)
}
