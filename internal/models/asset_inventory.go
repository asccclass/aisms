package models

import "time"

// AssetInventoryRecord ISMS-04-008 資訊資產清冊資料
type AssetInventoryRecord struct {
	ID                         int       `json:"id"`
	SystemName                 string    `json:"system_name"`
	AssetCode                  string    `json:"asset_code"`
	AssetType                  string    `json:"asset_type"`
	AssetName                  string    `json:"asset_name"`
	VendorName                 string    `json:"vendor_name"`
	IsCoreAsset                string    `json:"is_core_asset"`
	HasNationalSecurityConcern string    `json:"has_national_security_concern"`
	AssetDescription           string    `json:"asset_description"`
	Quantity                   string    `json:"quantity"`
	OsConfigBaseline           string    `json:"os_config_baseline"`
	BrowserConfigBaseline      string    `json:"browser_config_baseline"`
	NetworkConfigBaseline      string    `json:"network_config_baseline"`
	ApplicationConfigBaseline  string    `json:"application_config_baseline"`
	OtherConfigBaseline        string    `json:"other_config_baseline"`
	ConfigExceptionCode        string    `json:"config_exception_code"`
	ManagerDepartment          string    `json:"manager_department"`
	UserDepartment             string    `json:"user_department"`
	Location                   string    `json:"location"`
	Confidentiality            string    `json:"confidentiality"`
	Integrity                  string    `json:"integrity"`
	Availability               string    `json:"availability"`
	AssetValue                 string    `json:"asset_value"`
	LegalCompliance            string    `json:"legal_compliance"`
	ProtectionLevel            string    `json:"protection_level"`
	Mtpd                       string    `json:"mtpd"`
	Rto                        string    `json:"rto"`
	Rpo                        string    `json:"rpo"`
	Status                     string    `json:"status"`
	Creator                    string    `json:"creator"`
	Remarks                    string    `json:"remarks"`
	CreatedAt                  time.Time `json:"created_at"`
	UpdatedAt                  time.Time `json:"updated_at"`
}
