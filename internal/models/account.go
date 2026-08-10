package models

import "time"

// AccountStatus 帳號狀態
type AccountStatus string

const (
	StatusActive   AccountStatus = "active"   // 使用中
	StatusClosed   AccountStatus = "closed"   // 關閉
	StatusPending  AccountStatus = "pending"  // 待確認
	StatusExpired  AccountStatus = "expired"  // 已過期
)

// AccountType 帳號種類
type AccountType string

const (
	TypeDefault AccountType = "預設" // 預設帳號
	TypeSysAdmin AccountType = "系統管理用"
	TypeClosed   AccountType = "關閉帳號"
)

// PrivilegedAccount ISMS 特殊權限帳號
type PrivilegedAccount struct {
	ID              int           `json:"id"`
	SystemName      string        `json:"system_name"`      // 系統名稱
	Environment     string        `json:"environment"`      // 類別：開發區/測試區/正式區
	IPAddress       string        `json:"ip_address"`       // IP位址
	InventoryDate   string        `json:"inventory_date"`   // 盤點日期
	AccountName     string        `json:"account_name"`     // 帳號名稱
	AccountType     string        `json:"account_type"`     // 帳號種類
	DepartmentCode  string        `json:"department_code"`  // 單位代碼
	Department      string        `json:"department"`       // 單位名稱
	Creator         string        `json:"creator"`          // 輸入人員
	OwnerName       string        `json:"owner_name"`       // 姓名
	Email           string        `json:"email"`            // 電子郵件
	PassphraseRotate string       `json:"passphrase_rotate"` // 通行碼是否定期變更
	Status          AccountStatus `json:"status"`           // 狀態
	Remarks         string        `json:"remarks"`          // 備註
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	LastConfirmedAt *time.Time    `json:"last_confirmed_at"` // 最後確認時間
	ConfirmToken    string        `json:"-"`                 // 確認 token
	TokenExpiry     *time.Time    `json:"-"`                 // Token 過期時間
}

// ConfirmRequest 使用者確認請求
type ConfirmRequest struct {
	ID       int    `json:"id"`
	Token    string `json:"token"`
	Action   string `json:"action"` // "continue" or "stop"
	Comments string `json:"comments"`
}

// NotificationLog Email 發送紀錄
type NotificationLog struct {
	ID          int       `json:"id"`
	AccountID   int       `json:"account_id"`
	AccountName string    `json:"account_name"`
	Email       string    `json:"email"`
	SentAt      time.Time `json:"sent_at"`
	Status      string    `json:"status"` // "sent", "failed"
	Message     string    `json:"message"`
}

// DashboardFocusItems 儀表板重點提醒文案
type DashboardFocusItems struct {
	ActiveTitle  string `json:"active_title"`
	ActiveMeta   string `json:"active_meta"`
	PendingTitle string `json:"pending_title"`
	PendingMeta  string `json:"pending_meta"`
	ClosedTitle  string `json:"closed_title"`
	ClosedMeta   string `json:"closed_meta"`
	RecentTitle  string `json:"recent_title"`
}

// DashboardForm ISMS 表單註冊設定
type DashboardForm struct {
	ID                       int                 `json:"id"`
	Key                      string              `json:"key"`
	Code                     string              `json:"code"`
	ShortCode                string              `json:"short_code"`
	Name                     string              `json:"name"`
	Description              string              `json:"description"`
	DetailTitle              string              `json:"detail_title"`
	EmptyText                string              `json:"empty_text"`
	StatusNormalText         string              `json:"status_normal_text"`
	StatusNeedsAttentionText string              `json:"status_needs_attention_text"`
	ProviderKey              string              `json:"provider_key"`
	DisplayOrder             int                 `json:"display_order"`
	Enabled                  bool                `json:"enabled"`
	FocusItems               DashboardFocusItems `json:"focus_items"`
	CreatedAt                time.Time           `json:"created_at"`
	UpdatedAt                time.Time           `json:"updated_at"`
}

// DashboardFormProvider 表單資料來源選項
type DashboardFormProvider struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// DashboardRecord 首頁儀表板通用資料格式
//
// Provider 開發指南：
// 1. 每個 provider 最終都要把自己的資料表模型轉成 DashboardRecord。
// 2. 前端首頁只依賴這個格式，因此新增 provider 時通常不需要改首頁渲染。
// 3. 建議對應關係如下：
//    - PrimaryName: 卡片/清單主標題，例如帳號名稱、表單標題、資產名稱
//    - SecondaryName: 次要分類，例如系統/環境、類別、部門
//    - OwnerName: 負責人或使用者
//    - Status: active / pending / closed / expired 等首頁既有狀態值
//    - InventoryDate: 盤點日或主要業務日期，純字串即可
//    - UpdatedAt: 用於首頁排序與最近更新顯示
type DashboardRecord struct {
	ID            int       `json:"id"`
	PrimaryName   string    `json:"primary_name"`
	SecondaryName string    `json:"secondary_name"`
	OwnerName     string    `json:"owner_name"`
	Status        string    `json:"status"`
	InventoryDate string    `json:"inventory_date"`
	UpdatedAt     time.Time `json:"updated_at"`
	Email         string    `json:"email"`
}

// CustomTableTemplateRecord 自訂資料表 provider 範本資料模型
//
// 這是一個「照抄即可改」的 provider 範本模型：
// - 若要接第二張真表，可複製這個 struct 並改名
// - 接著在 db 層新增 ListXxxRecords()
// - 最後在 handlers.go 的 loadDashboardRecords() switch 裡補一個 case
type CustomTableTemplateRecord struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Category      string    `json:"category"`
	OwnerName     string    `json:"owner_name"`
	Status        string    `json:"status"`
	InventoryDate string    `json:"inventory_date"`
	Email         string    `json:"email"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
