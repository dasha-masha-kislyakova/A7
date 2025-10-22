package internal

import "time"

type Application struct {
	ID             int64 `json:"id"`
	PickupPointID  int64 `json:"pickup_point_id"`
	DropoffPointID int64 `json:"dropoff_point_id"`

	// 1) Данные заказчика
	SendNameCompany string `json:"send_name_company" validate:"required"`
	SendInnCompany  string `json:"send_inn_company" validate:"required"`
	SendFIO         string `json:"fio" validate:"required"`
	SendTelNumber   string `json:"tel_number" validate:"required"`
	SendEmail       string `json:"email" validate:"required"`

	// 2) Данные о грузу
	CargoName           string  `json:"cargo_name" validate:"required"`
	Boxes               int64   `json:"boxes" validate:"required"`
	TotalWeight         float64 `json:"total_weight" validate:"required"`
	TotalValume         float64 `json:"total_valume" validate:"required"`
	SpecialRequirements string  `json:"special_requirements,omitempty"`

	// 3) Данные о получателе
	RecipNameCompany string `json:"recip_name_company" validate:"required"`
	RecipAddress     string `json:"recip_address" validate:"required"`
	RecipFIO         string `json:"recip_fio" validate:"required"`
	RecipTelephone   string `json:"recip_telephone" validate:"required"`

	// 4) Служебные
	Status    string    `json:"status" validate:"omitempty,oneof=PENDING APPROVED"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
	AcceptedAt         *time.Time        `json:"accepted_at,omitempty"`
	DispatchedAt       *time.Time        `json:"dispatched_at,omitempty"`
	DeliveredAt        *time.Time        `json:"delivered_at,omitempty"`
}
}
