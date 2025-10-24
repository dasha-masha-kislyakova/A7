package office

import (
	"context"
	"errors"
)

type Service interface {
	Create(ctx context.Context, req CreateApplicationRequest) (Application, error)
	Get(ctx context.Context, id int64) (Application, error)
	UpdateStatus(ctx context.Context, id int64, status ApplicationStatus) error
	List(ctx context.Context, status *string) ([]Application, error)
}

type service struct{ repo Repo }

func NewService(repo Repo) (Service, error) {
	if repo == nil {
		return nil, errors.New("nil repo")
	}
	return &service{repo: repo}, nil
}

func (s *service) Create(ctx context.Context, req CreateApplicationRequest) (Application, error) {
	// базовая валидация (по ТЗ — все обязательные поля) :contentReference[oaicite:1]{index=1}
	if req.LogisticsPointID == 0 ||
		req.SenderOrgName == "" || req.SenderINN == "" ||
		req.SenderContactFIO == "" || req.SenderContactPhone == "" ||
		req.CargoName == "" || req.CargoCount <= 0 || req.CargoWeight <= 0 || req.CargoVolume <= 0 ||
		req.RecipientOrgName == "" || req.RecipientAddress == "" || req.RecipientContactFIO == "" || req.RecipientContactPhone == "" {
		return Application{}, errors.New("missing required fields")
	}
	return s.repo.Insert(ctx, req)
}

func (s *service) Get(ctx context.Context, id int64) (Application, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) UpdateStatus(ctx context.Context, id int64, status ApplicationStatus) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *service) List(ctx context.Context, status *string) ([]Application, error) {
	return s.repo.List(ctx, status)
}
