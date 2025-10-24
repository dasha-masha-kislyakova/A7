package audit

import "time"

type Event struct {
	ID            int64     `json:"id"`
	SourceService string    `json:"source_service"` // кто запросил
	TargetService string    `json:"target_service"` // кто обслужил
	URI           string    `json:"uri"`
	HTTPStatus    int       `json:"http_status"`
	At            time.Time `json:"at"` // дата события
	DurationMs    int64     `json:"duration_ms"`
	UserID        string    `json:"user_id,omitempty"`
	RequestBody   string    `json:"request_body,omitempty"`
	ResponseBody  string    `json:"response_body,omitempty"`
}

type Query struct {
	SourceService string
	TargetService string
	URI           string
	HTTPStatus    *int
	DateFrom      *time.Time
	DateTo        *time.Time
	MinDurationMs *int64
	UserID        string

	SortBy    string // source|target|status|duration|user
	SortOrder string // asc|desc
	Page      int
	PageSize  int // 10|50|100
}
