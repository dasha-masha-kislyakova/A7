package internal

import "time"

// Привязка заявки к перевозке. Веса/объёмы фиксируем с заявки на момент назначения.
type Assignment struct {
	ID                     int64     `json:"id"`
	ShipmentID             int64     `json:"shipment_id"`
	ApplicationExternalID  int64     `json:"application_external_id"`             // id заявки
	PickupPointExternalID  *int64    `json:"pickup_point_external_id,omitempty"`  // опционально — для аналитики
	DropoffPointExternalID *int64    `json:"dropoff_point_external_id,omitempty"` // опционально — для аналитики
	WeightKg               float64   `json:"weight_kg"`
	VolumeM3               float64   `json:"volume_m3"`
	CreatedAt              time.Time `json:"created_at"`
}
