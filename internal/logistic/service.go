package logistic

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Service interface {
	GetLogApp(ctx context.Context, id int64) (LogisticApplication, error)
	ListLogApps(ctx context.Context, status *string) ([]LogisticApplication, error)
	UpdateLogAppStatus(ctx context.Context, id int64, status ApplicationStatus) error

	CreateRoute(ctx context.Context, req CreateRouteRequest) (Route, error)
	AssignApp(ctx context.Context, routeID, originalAppID int64) error
	SendRoute(ctx context.Context, routeID int64) error
}

type service struct {
	repo     Repo
	officeCB string
}

func NewService(repo Repo, officeInternalBaseURL string) (Service, error) {
	if repo == nil {
		return nil, errors.New("nil repo")
	}
	return &service{repo: repo, officeCB: strings.TrimRight(officeInternalBaseURL, "/")}, nil
}

func (s *service) GetLogApp(ctx context.Context, id int64) (LogisticApplication, error) {
	return s.repo.GetLogApp(ctx, id)
}

func (s *service) ListLogApps(ctx context.Context, status *string) ([]LogisticApplication, error) {
	return s.repo.ListLogApps(ctx, status)
}

func (s *service) UpdateLogAppStatus(ctx context.Context, id int64, status ApplicationStatus) error {
	if err := s.repo.UpdateLogAppStatus(ctx, id, status); err != nil {
		return err
	}
	app, err := s.repo.GetLogApp(ctx, id)
	if err != nil {
		return err
	}
	body, _ := json.Marshal(struct {
		Status ApplicationStatus `json:"status"`
	}{Status: status})
	req, _ := http.NewRequestWithContext(ctx, "POST",
		s.officeCB+"/office/applications/"+strconv.FormatInt(app.OriginalApplicationID, 10)+"/status",
		strings.NewReader(string(body)),
	)
	req.Header.Set("Content-Type", "application/json")
	_, _ = http.DefaultClient.Do(req)
	return nil
}

func (s *service) CreateRoute(ctx context.Context, req CreateRouteRequest) (Route, error) {
	if req.TruckVolume <= 0 || req.TruckMaxWeight <= 0 || len(req.RoutePoints) < 2 {
		return Route{}, errors.New("invalid route data")
	}
	route, err := s.repo.InsertRoute(ctx, req)
	if err != nil {
		return Route{}, err
	}
	for _, p := range req.RoutePoints {
		if err := s.repo.InsertRoutePoint(ctx, route.ID, p); err != nil {
			return Route{}, err
		}
	}
	return route, nil
}

func (s *service) AssignApp(ctx context.Context, routeID, originalAppID int64) error {
	logAppID, err := s.repo.FindOrCreateLogApp(ctx, originalAppID)
	if err != nil {
		return err
	}
	return s.repo.AssignRouteApp(ctx, routeID, originalAppID, logAppID)
}

func (s *service) SendRoute(ctx context.Context, routeID int64) error {
	if err := s.repo.SetRouteInProgress(ctx, routeID); err != nil {
		return err
	}
	pairs, err := s.repo.RouteAppPairs(ctx, routeID)
	if err != nil {
		return err
	}
	for _, p := range pairs {
		logID, officeID := p[0], p[1]
		_ = s.repo.UpdateLogAppStatus(ctx, logID, StatusInProgress)

		body, _ := json.Marshal(struct {
			Status ApplicationStatus `json:"status"`
		}{Status: StatusInProgress})
		req, _ := http.NewRequestWithContext(ctx, "POST",
			s.officeCB+"/office/applications/"+strconv.FormatInt(officeID, 10)+"/status",
			strings.NewReader(string(body)),
		)
		req.Header.Set("Content-Type", "application/json")
		_, _ = http.DefaultClient.Do(req)
	}
	return nil
}
