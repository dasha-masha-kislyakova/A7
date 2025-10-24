package office

import "time"

type ApplicationStatus string

const (
	StatusNew        ApplicationStatus = "NEW"
	StatusInProgress ApplicationStatus = "IN_PROGRESS"
	StatusShipped    ApplicationStatus = "SHIPPED"
	StatusDelivered  ApplicationStatus = "DELIVERED"
	StatusCancelled  ApplicationStatus = "CANCELLED"
)

type Application struct {
	ID               int64             `json:"id"`
	Status           ApplicationStatus `json:"status"`
	LogisticsPointID int64             `json:"logistics_point_id"`

	SenderOrgName      string  `json:"sender_org_name"`
	SenderINN          string  `json:"sender_inn"`
	SenderContactFIO   string  `json:"sender_contact_fio"`
	SenderContactPhone string  `json:"sender_contact_phone"`
	SenderEmail        *string `json:"sender_email,omitempty"`

	CargoName           string  `json:"cargo_name"`
	CargoCount          int     `json:"cargo_count"`
	CargoWeight         float64 `json:"cargo_weight"`
	CargoVolume         float64 `json:"cargo_volume"`
	SpecialRequirements *string `json:"special_requirements,omitempty"`

	RecipientOrgName      string `json:"recipient_org_name"`
	RecipientAddress      string `json:"recipient_address"`
	RecipientContactFIO   string `json:"recipient_contact_fio"`
	RecipientContactPhone string `json:"recipient_contact_phone"`

	CreatedByManagerID int64     `json:"created_by_manager_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type CreateApplicationRequest struct {
	LogisticsPointID      int64   `json:"logistics_point_id"`
	SenderOrgName         string  `json:"sender_org_name"`
	SenderINN             string  `json:"sender_inn"`
	SenderContactFIO      string  `json:"sender_contact_fio"`
	SenderContactPhone    string  `json:"sender_contact_phone"`
	SenderEmail           *string `json:"sender_email"`
	CargoName             string  `json:"cargo_name"`
	CargoCount            int     `json:"cargo_count"`
	CargoWeight           float64 `json:"cargo_weight"`
	CargoVolume           float64 `json:"cargo_volume"`
	SpecialRequirements   *string `json:"special_requirements"`
	RecipientOrgName      string  `json:"recipient_org_name"`
	RecipientAddress      string  `json:"recipient_address"`
	RecipientContactFIO   string  `json:"recipient_contact_fio"`
	RecipientContactPhone string  `json:"recipient_contact_phone"`
}

type UpdateStatusRequest struct {
	Status ApplicationStatus `json:"status"`
}
