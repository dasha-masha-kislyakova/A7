package internal

import "time"

type LogisticPoint struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Address   *string   `json:"address,omitempty"`
	Lat       *float64  `json:"lat,omitempty"`
	Lon       *float64  `json:"lon,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
