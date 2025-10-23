package logistic

import "time"

type ShipmentStatus string

const (
	ShipPLANNED   ShipmentStatus = "PLANNED"
	ShipDEPARTED  ShipmentStatus = "DEPARTED"
	ShipDELIVERED ShipmentStatus = "DELIVERED"
)

type Shipment struct {
	ID          int64          `json:"id"`
	MaxWeightKg float64        `json:"max_weight_kg"`
	MaxVolumeM3 float64        `json:"max_volume_m3"`
	DepartureAt time.Time      `json:"departure_at"`
	Status      ShipmentStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
}

type RoutePoint struct {
	ID              int64     `json:"id"`
	ShipmentID      int64     `json:"shipment_id"`
	PointID         int64     `json:"point_id"`
	PlannedArriveAt time.Time `json:"planned_arrive_at"`
	Ordinal         int16     `json:"ordinal"`
}

type LogisticPoint struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Address   *string   `json:"address,omitempty"`
	Lat       *float64  `json:"lat,omitempty"`
	Lon       *float64  `json:"lon,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Assignment struct {
	ID                     int64     `json:"id"`
	ShipmentID             int64     `json:"shipment_id"`
	ApplicationExternalID  int64     `json:"application_external_id"`
	PickupPointExternalID  *int64    `json:"pickup_point_external_id,omitempty"`
	DropoffPointExternalID *int64    `json:"dropoff_point_external_id,omitempty"`
	WeightKg               float64   `json:"weight_kg"`
	VolumeM3               float64   `json:"volume_m3"`
	CreatedAt              time.Time `json:"created_at"`
}
