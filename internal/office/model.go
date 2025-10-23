package office

import "time"

type ApplicationStatus string

const (
	AppNEW       ApplicationStatus = "NEW"
	AppINWORK    ApplicationStatus = "IN_WORK"
	AppINTRANSIT ApplicationStatus = "IN_TRANSIT"
	AppDELIVERED ApplicationStatus = "DELIVERED"
)

type Client struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	INN   string `json:"inn"`
	FIO   string `json:"fio"`
	Tel   string `json:"tel"` // телефон как строка (поддержка +, пробелов и пр.)
	Email string `json:"email"`
}

type Cargo struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Boxes       int64   `json:"boxes"`
	TotalWeight float64 `json:"total_weight"`
	TotalVolume float64 `json:"total_volume"`
	SpecialReq  *string `json:"special_requirements,omitempty"`
}

type Recipient struct {
	ID        int64  `json:"id"`
	Name      string `json:"name_company"`
	Address   string `json:"address"`
	FIO       string `json:"fio"`
	Telephone string `json:"telephone"` // как строка
}

type ApplicationRow struct {
	ID             int64
	PickupPointID  int64
	DropoffPointID int64
	ClientID       int64
	CargoID        int64
	RecipientID    int64
	Status         string
	CreatedAt      time.Time
	AcceptedAt     *time.Time
	DispatchedAt   *time.Time
	DeliveredAt    *time.Time
}
