package internal

import "time"

type transfer struct {
	ID        int64   `json:"id"`
	MaxWeight float64 `json:"max_weight"`
	MaxVolume float64 `json:"max_volume"`
	Data      time.Time
	Status    string `json:"status"`
}
