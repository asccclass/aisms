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
