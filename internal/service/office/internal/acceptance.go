package internal

import "time"

type ApplicationAcceptance struct {
	ID                    int64     `json:"id"`
	ApplicationExternalID int64     `json:"application_external_id"`
	AcceptedAt            time.Time `json:"accepted_at"`
	AcceptedByUserID      int64     `json:"accepted_by_user_id"` // админ офиса
	Note                  *string   `json:"note,omitempty"`
}
