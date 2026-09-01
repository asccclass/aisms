package models

// CreatorBackfillResult 回填舊資料歸屬的結果
type CreatorBackfillResult struct {
	PrivilegedAccounts   int `json:"privileged_accounts"`
	SystemPlatforms      int `json:"system_platform_requests"`
	FirewallRequests     int `json:"firewall_requests"`
	AssetInventory       int `json:"asset_inventory_records"`
	ProtectionBaselines  int `json:"protection_baseline_records"`
	TotalUpdated         int `json:"total_updated"`
}
