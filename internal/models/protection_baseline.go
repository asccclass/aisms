package models

import "time"

// ProtectionBaselineControl 04-069 控制措施明細
type ProtectionBaselineControl struct {
	ID                  int    `json:"id"`
	RecordID            int    `json:"record_id"`
	ItemNo              int    `json:"item_no"`
	DomainName          string `json:"domain_name"`
	ControlCategory     string `json:"control_category"`
	RequirementLevel    string `json:"requirement_level"`
	ControlDescription  string `json:"control_description"`
	MeasureNotes        string `json:"measure_notes"`
	Applies             string `json:"applies"`
	ImplementationNotes string `json:"implementation_notes"`
	Compliance          string `json:"compliance"`
	Finding             string `json:"finding"`
	Remarks             string `json:"remarks"`
}

// ProtectionBaselineRecord 04-069 主表
type ProtectionBaselineRecord struct {
	ID             int                         `json:"id"`
	SystemName     string                      `json:"system_name"`
	SecurityLevel  string                      `json:"security_level"`
	FilledBy       string                      `json:"filled_by"`
	FormDate       string                      `json:"form_date"`
	Reviewer       string                      `json:"reviewer"`
	ReviewDate     string                      `json:"review_date"`
	Status         string                      `json:"status"`
	Creator        string                      `json:"creator"`
	Remarks        string                      `json:"remarks"`
	Controls       []ProtectionBaselineControl `json:"controls"`
	TotalControls  int                         `json:"total_controls"`
	AppliedCount   int                         `json:"applied_count"`
	CompliantCount int                         `json:"compliant_count"`
	CreatedAt      time.Time                   `json:"created_at"`
	UpdatedAt      time.Time                   `json:"updated_at"`
}
