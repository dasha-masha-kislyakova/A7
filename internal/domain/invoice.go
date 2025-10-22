package domain

import "time"

type Application struct {
	ID               int64   `json:"id"`
	PickupPointID    int64   `json:"pickup_point_id"`
	DropoffPointID   int64   `json:"dropoff_point_id"`

	// 1) Данные заказчика
	CompanyName      string  `json:"company_name"`
	CompanyINN       string  `json:"company_inn"`
	FIO              string  `json:"fio"`
	TelephoneNumber  string  `json:"telephone_number"`
	Email            string  `json:"email"`

	// 2) Данные о грузу
	CargoName        string  `json:"cargo_name"`
	Boxes           int64  `json:"boxes"`
	TotalWeight      float64 `json:"total_weight"`
	TotalValume     float64 `json:"total_valume"`
	SpecialRequirements string  `json:"special_requirements,omitempty"`

	// 3) Данные о получателе
	RecipNameCompany   string  `json:"recip_name_company,omitempty"`
