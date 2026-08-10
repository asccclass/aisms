package models

import "time"

// SystemPlatformRequest ISMS-04-078 系統平台申請資料
type SystemPlatformRequest struct {
	ID                   int       `json:"id"`
	RequestDate          string    `json:"request_date"`
	ApplicantName        string    `json:"applicant_name"`
	ApplicantDepartment  string    `json:"applicant_department"`
	ApplicantTitle       string    `json:"applicant_title"`
	OfficePhone          string    `json:"office_phone"`
	Email                string    `json:"email"`
	PIName               string    `json:"pi_name"`
	SystemName           string    `json:"system_name"`
	SystemAlias          string    `json:"system_alias"`
	SystemPurpose        string    `json:"system_purpose"`
	EstimatedUsers       string    `json:"estimated_users"`
	InternalOnly         string    `json:"internal_only"`
	IPRestriction        string    `json:"ip_restriction"`
	RequestStartDate     string    `json:"request_start_date"`
	RequestEndDate       string    `json:"request_end_date"`
	RequestType          string    `json:"request_type"`
	ShutdownRetainMonths string    `json:"shutdown_retain_months"`
	ShutdownReason       string    `json:"shutdown_reason"`
	EnvironmentType      string    `json:"environment_type"`
	OperatingSystem      string    `json:"operating_system"`
	OperatingSystemOther string    `json:"operating_system_other"`
	DiskSize             string    `json:"disk_size"`
	SpecialRequirements  string    `json:"special_requirements"`
	DomainSettings       string    `json:"domain_settings"`
	OtherRequirements    string    `json:"other_requirements"`
	BackupRequired       string    `json:"backup_required"`
	BackupRequirements   string    `json:"backup_requirements"`
	BackupReason         string    `json:"backup_reason"`
	ApplicantSignature   string    `json:"applicant_signature"`
	SupervisorSignature  string    `json:"supervisor_signature"`
	Status               string    `json:"status"`
	Creator              string    `json:"creator"`
	Remarks              string    `json:"remarks"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
