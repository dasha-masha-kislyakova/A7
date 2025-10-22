package internal

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
	Route       []RoutePoint   `json:"route"` // ровно 2 точки (ordinal 0 и 1)
}

type RoutePoint struct {
	ID              int64     `json:"id"`
	ShipmentID      int64     `json:"shipment_id"`
	Title           string    `json:"title"`
	Address         *string   `json:"address,omitempty"`
	Lat             *float64  `json:"lat,omitempty"`
	Lon             *float64  `json:"lon,omitempty"`
	PlannedArriveAt time.Time `json:"planned_arrive_at"`
	Ordinal         int16     `json:"ordinal"` // 0 или 1
}
