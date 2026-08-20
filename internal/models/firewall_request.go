package models

import "time"

// FirewallRequest ISMS-04-042 防火牆申請單資料
type FirewallRequest struct {
	ID               int       `json:"id"`
	LegacyFormNumber string    `json:"legacy_form_number"`
	SystemName       string    `json:"system_name"`
	Action           string    `json:"action"`
	PurposeType      string    `json:"purpose_type"`
	SourceZone       string    `json:"source_zone"`
	SourceZone2      string    `json:"source_zone2"`
	SourceIP         string    `json:"source_ip"`
	DestinationZone  string    `json:"destination_zone"`
	DestinationZone2 string    `json:"destination_zone2"`
	DestinationIP    string    `json:"destination_ip"`
	ProtocolType     string    `json:"protocol_type"`
	StartDate        string    `json:"start_date"`
	EndDate          string    `json:"end_date"`
	RequestDate      string    `json:"request_date"`
	RuleDescription  string    `json:"rule_description"`
	FirewallZone     string    `json:"firewall_zone"`
	FirewallID       string    `json:"firewall_id"`
	Status           string    `json:"status"`
	Creator          string    `json:"creator"`
	Remarks          string    `json:"remarks"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
