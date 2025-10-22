package domain

import "time"

type LogisticPoint struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Volume    float64   `json:"volume"`
	Weight    float64   `json:"weight"`
	Address   *string   `json:"address,omitempty"`
	Lat       *float64  `json:"lat,omitempty"`
	Lon       *float64  `json:"lon,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
