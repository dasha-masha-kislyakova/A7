package service

import (
	"a7/internal/office"
	"a7/internal/office/repo"
	"strings"
	"time"
)

type Service struct{ R *repo.Repo }

func New(r *repo.Repo) *Service { return &Service{R: r} }

func (s *Service) CreateApplication(pp, dp int64, c office.Client, g office.Cargo, rc office.Recipient) (int64, error) {
	if pp == 0 || dp == 0 || pp == dp {
		return 0, ErrTwoDistinctPoints
	}
	c.Name = strings.TrimSpace(c.Name)
	c.INN = strings.TrimSpace(c.INN)
	cid, err := s.R.UpsertClient(&c)
	if err != nil {
		return 0, err
	}
	gid, err := s.R.InsertCargo(&g)
	if err != nil {
		return 0, err
	}
	reid, err := s.R.InsertRecipient(&rc)
	if err != nil {
		return 0, err
	}
	return s.R.InsertApplication(&office.ApplicationRow{PickupPointID: pp, DropoffPointID: dp, ClientID: cid, CargoID: gid, RecipientID: reid, Status: string(office.AppNEW), CreatedAt: time.Now()})
}

func (s *Service) Accept(id int64) error  { return s.R.ChangeStatus(id, string(office.AppINWORK)) }
func (s *Service) Deliver(id int64) error { return s.R.ChangeStatus(id, string(office.AppDELIVERED)) }
func (s *Service) MarkInTransit(id int64) error {
	return s.R.ChangeStatus(id, string(office.AppINTRANSIT))
}

func (s *Service) GetApplication(id int64) (map[string]any, error) { return s.R.GetApplication(id) }
func (s *Service) ListApplications(status string) ([]map[string]any, error) {
	return s.R.ListApplications(status)
}
func (s *Service) ListAvailableForRoute(pp, dp int64, before time.Time) ([]map[string]any, error) {
	return s.R.ListAvailableForRoute(pp, dp, before)
}

var ErrTwoDistinctPoints = &AppError{Msg: "должно быть указано ровно 2 разные точки"}

type AppError struct{ Msg string }

func (e *AppError) Error() string { return e.Msg }
