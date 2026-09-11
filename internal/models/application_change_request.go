package models

import "time"

// ApplicationChangeRequest ISMS-04-052 應用系統功能需求更新建議單資料
type ApplicationChangeRequest struct {
	ID                        int       `json:"id"`
	Suggestor                 string    `json:"suggestor"`
	FormDate                  string    `json:"form_date"`
	Approver                  string    `json:"approver"`
	SystemName                string    `json:"system_name"`
	FeatureName               string    `json:"feature_name"`
	IsRequiredFeature         string    `json:"is_required_feature"`
	IsMajorImpact             string    `json:"is_major_impact"`
	ExpectedOnlineDate        string    `json:"expected_online_date"`
	BackgroundDescription     string    `json:"background_description"`
	ExistingSecurityMeasures  string    `json:"existing_security_measures"`
	SecurityScopeInvolved     string    `json:"security_scope_involved"`
	AccessControlMeasures     string    `json:"access_control_measures"`
	AuditMeasures             string    `json:"audit_measures"`
	ContinuityMeasures        string    `json:"continuity_measures"`
	IdentificationMeasures    string    `json:"identification_measures"`
	SystemAcquisitionMeasures string    `json:"system_acquisition_measures"`
	CommunicationMeasures     string    `json:"communication_measures"`
	IntegrityMeasures         string    `json:"integrity_measures"`
	CapacityManagement        string    `json:"capacity_management"`
	CapacityChangeDescription string    `json:"capacity_change_description"`
	OtherSecurityMeasures     string    `json:"other_security_measures"`
	OtherSecurityDescription  string    `json:"other_security_description"`
	InformationServiceOpinion string    `json:"information_service_opinion"`
	MeetingTime               string    `json:"meeting_time"`
	MeetingDecision           string    `json:"meeting_decision"`
	RejectionReason           string    `json:"rejection_reason"`
	CoordinatingStaff         string    `json:"coordinating_staff"`
	CoordinationDate          string    `json:"coordination_date"`
	CoordinationApprover      string    `json:"coordination_approver"`
	Status                    string    `json:"status"`
	Creator                   string    `json:"creator"`
	Remarks                   string    `json:"remarks"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}
