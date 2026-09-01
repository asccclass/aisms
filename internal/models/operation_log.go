package models

import "time"

// OperationLog 系統操作日誌
type OperationLog struct {
	ID           int       `json:"id"`
	EventType    string    `json:"event_type"`
	OccurredAt   time.Time `json:"occurred_at"`
	Location     string    `json:"location"`
	RequestPath  string    `json:"request_path"`
	RequestMethod string   `json:"request_method"`
	StatusCode   int       `json:"status_code"`
	SourceIP     string    `json:"source_ip"`
	UserAgent    string    `json:"user_agent"`
	UserEmail    string    `json:"user_email"`
	UserName     string    `json:"user_name"`
	UserGoogleID string    `json:"user_google_id"`
	Department   string    `json:"department"`
}
