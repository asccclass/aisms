package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"isms-privilege/internal/models"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

// New 建立資料庫連線並初始化
func New(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

func (d *DB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS privileged_accounts (
		id               INTEGER PRIMARY KEY AUTOINCREMENT,
		system_name      TEXT    NOT NULL,
		environment      TEXT    NOT NULL DEFAULT '正式區',
		ip_address       TEXT    NOT NULL,
		inventory_date   TEXT    NOT NULL,
		account_name     TEXT    NOT NULL,
		account_type     TEXT    NOT NULL DEFAULT '系統管理用',
		department_code  TEXT    NOT NULL DEFAULT '',
		department       TEXT    NOT NULL DEFAULT '',
		creator          TEXT    NOT NULL DEFAULT '',
		owner_name       TEXT    NOT NULL DEFAULT '',
		email            TEXT    NOT NULL DEFAULT '',
		passphrase_rotate TEXT   NOT NULL DEFAULT '是',
		status           TEXT    NOT NULL DEFAULT 'active',
		remarks          TEXT    NOT NULL DEFAULT '',
		created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_confirmed_at DATETIME,
		confirm_token    TEXT,
		token_expiry     DATETIME
	);

	CREATE TABLE IF NOT EXISTS notification_logs (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		account_id   INTEGER NOT NULL REFERENCES privileged_accounts(id),
		account_name TEXT    NOT NULL,
		email        TEXT    NOT NULL,
		sent_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		status       TEXT    NOT NULL DEFAULT 'sent',
		message      TEXT    NOT NULL DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS dashboard_forms (
		id                           INTEGER PRIMARY KEY AUTOINCREMENT,
		form_key                     TEXT    NOT NULL UNIQUE,
		code                         TEXT    NOT NULL,
		short_code                   TEXT    NOT NULL,
		name                         TEXT    NOT NULL,
		description                  TEXT    NOT NULL DEFAULT '',
		detail_title                 TEXT    NOT NULL DEFAULT '最近資料清單',
		empty_text                   TEXT    NOT NULL DEFAULT '此表單尚未接入資料來源',
		status_normal_text           TEXT    NOT NULL DEFAULT '狀況正常',
		status_needs_attention_text  TEXT    NOT NULL DEFAULT '需追蹤',
		provider_key                 TEXT    NOT NULL DEFAULT 'placeholder',
		display_order                INTEGER NOT NULL DEFAULT 0,
		enabled                      INTEGER NOT NULL DEFAULT 1,
		focus_items_json             TEXT    NOT NULL DEFAULT '{}',
		created_at                   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at                   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS custom_table_template_records (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		title          TEXT    NOT NULL,
		category       TEXT    NOT NULL DEFAULT '',
		owner_name     TEXT    NOT NULL DEFAULT '',
		status         TEXT    NOT NULL DEFAULT 'active',
		inventory_date TEXT    NOT NULL DEFAULT '',
		email          TEXT    NOT NULL DEFAULT '',
		created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := d.conn.Exec(schema)
	_, _ = d.conn.Exec(`ALTER TABLE privileged_accounts ADD COLUMN environment TEXT NOT NULL DEFAULT '正式區'`)
	_, _ = d.conn.Exec(`ALTER TABLE privileged_accounts ADD COLUMN department_code TEXT NOT NULL DEFAULT ''`)
	_, _ = d.conn.Exec(`ALTER TABLE privileged_accounts ADD COLUMN creator TEXT NOT NULL DEFAULT ''`)
	if err == nil {
		_ = d.seedDashboardForms()
		_ = d.seedCustomTableTemplateRecords()
	}
	return err
}

// seed 插入文件中的範例資料
func (d *DB) seed() {
	var count int
	d.conn.QueryRow("SELECT COUNT(*) FROM privileged_accounts").Scan(&count)
	if count > 0 {
		return
	}
	accounts := []models.PrivilegedAccount{
		{SystemName: "Sysldap", IPAddress: "10.109.4.9", InventoryDate: "20260803", AccountName: "root", AccountType: "預設", Department: "資安部", OwnerName: "顏景喆", Email: "ccyen@example.com", PassphraseRotate: "是", Status: models.StatusActive},
		{SystemName: "Sysldap", IPAddress: "10.109.4.9", InventoryDate: "20260803", AccountName: "ccyen", AccountType: "系統管理用", Department: "資安部", OwnerName: "顏景喆", Email: "ccyen@example.com", PassphraseRotate: "是", Status: models.StatusActive},
		{SystemName: "Sysldap", IPAddress: "10.109.4.9", InventoryDate: "20260803", AccountName: "ssh", AccountType: "系統管理用", Department: "資安部", OwnerName: "洪紹雄", Email: "ssh@example.com", PassphraseRotate: "是", Status: models.StatusActive},
		{SystemName: "Sysldap", IPAddress: "10.109.4.9", InventoryDate: "20260803", AccountName: "maxtung", AccountType: "系統管理用", Department: "資安部", OwnerName: "董君瀚", Email: "maxtung@example.com", PassphraseRotate: "是", Status: models.StatusActive},
		{SystemName: "Sysldap", IPAddress: "10.109.4.9", InventoryDate: "20260803", AccountName: "yuchi467", AccountType: "系統管理用", Department: "資安部", OwnerName: "蘇宥綺", Email: "yuchi467@example.com", PassphraseRotate: "是", Status: models.StatusActive},
		{SystemName: "Sysldap", IPAddress: "10.109.4.9", InventoryDate: "20260803", AccountName: "wzlu", AccountType: "系統管理用", Department: "資安部", OwnerName: "陸維肇", Email: "wzlu@example.com", PassphraseRotate: "是", Status: models.StatusActive},
		{SystemName: "Sysldap", IPAddress: "10.109.4.9", InventoryDate: "20260803", AccountName: "tedl", AccountType: "系統管理用", Department: "資安部", OwnerName: "廖鈺翔", Email: "tedl@example.com", PassphraseRotate: "是", Status: models.StatusActive},
		{SystemName: "Sysldap", IPAddress: "10.109.4.9", InventoryDate: "20260803", AccountName: "ktk", AccountType: "系統管理用", Department: "資安部", OwnerName: "高昆鈿", Email: "ktk@example.com", PassphraseRotate: "是", Status: models.StatusActive},
		{SystemName: "Sysldap", IPAddress: "10.109.4.9", InventoryDate: "20260803", AccountName: "leo24417", AccountType: "關閉帳號", Department: "", OwnerName: "離職同仁", Email: "", PassphraseRotate: "NA", Status: models.StatusClosed},
		{SystemName: "Sysldap", IPAddress: "10.109.4.9", InventoryDate: "20260803", AccountName: "morriswang", AccountType: "關閉帳號", Department: "", OwnerName: "職務異動", Email: "", PassphraseRotate: "NA", Status: models.StatusClosed},
		{SystemName: "Sysldap", IPAddress: "10.109.4.9", InventoryDate: "20260803", AccountName: "alexkau", AccountType: "關閉帳號", Department: "", OwnerName: "離職同仁", Email: "", PassphraseRotate: "NA", Status: models.StatusClosed},
	}
	for _, a := range accounts {
		d.conn.Exec(`INSERT INTO privileged_accounts (system_name,environment,ip_address,inventory_date,account_name,account_type,department_code,department,creator,owner_name,email,passphrase_rotate,status) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			a.SystemName, a.Environment, a.IPAddress, a.InventoryDate, a.AccountName, a.AccountType, a.DepartmentCode, a.Department, a.Creator, a.OwnerName, a.Email, a.PassphraseRotate, a.Status)
	}
}

// ---- CRUD ----

func (d *DB) ListAccounts(status string) ([]models.PrivilegedAccount, error) {
	q := `SELECT id,system_name,environment,ip_address,inventory_date,account_name,account_type,department_code,department,creator,owner_name,email,passphrase_rotate,status,remarks,created_at,updated_at,last_confirmed_at FROM privileged_accounts`
	args := []interface{}{}
	if status != "" && status != "all" {
		q += " WHERE status = ?"
		args = append(args, status)
	}
	q += " ORDER BY id DESC"
	rows, err := d.conn.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.PrivilegedAccount
	for rows.Next() {
		var a models.PrivilegedAccount
		var lc sql.NullTime
		if err := rows.Scan(&a.ID, &a.SystemName, &a.Environment, &a.IPAddress, &a.InventoryDate, &a.AccountName, &a.AccountType, &a.DepartmentCode, &a.Department, &a.Creator, &a.OwnerName, &a.Email, &a.PassphraseRotate, &a.Status, &a.Remarks, &a.CreatedAt, &a.UpdatedAt, &lc); err != nil {
			continue
		}
		if lc.Valid {
			a.LastConfirmedAt = &lc.Time
		}
		list = append(list, a)
	}
	return list, nil
}

func (d *DB) GetAccount(id int) (*models.PrivilegedAccount, error) {
	row := d.conn.QueryRow(`SELECT id,system_name,environment,ip_address,inventory_date,account_name,account_type,department_code,department,creator,owner_name,email,passphrase_rotate,status,remarks,created_at,updated_at,last_confirmed_at,confirm_token,token_expiry FROM privileged_accounts WHERE id=?`, id)
	var a models.PrivilegedAccount
	var lc, te sql.NullTime
	var tok sql.NullString
	if err := row.Scan(&a.ID, &a.SystemName, &a.Environment, &a.IPAddress, &a.InventoryDate, &a.AccountName, &a.AccountType, &a.DepartmentCode, &a.Department, &a.Creator, &a.OwnerName, &a.Email, &a.PassphraseRotate, &a.Status, &a.Remarks, &a.CreatedAt, &a.UpdatedAt, &lc, &tok, &te); err != nil {
		return nil, err
	}
	if lc.Valid {
		a.LastConfirmedAt = &lc.Time
	}
	if tok.Valid {
		a.ConfirmToken = tok.String
	}
	if te.Valid {
		a.TokenExpiry = &te.Time
	}
	return &a, nil
}

func (d *DB) GetAccountByToken(token string) (*models.PrivilegedAccount, error) {
	row := d.conn.QueryRow(`SELECT id,system_name,environment,ip_address,inventory_date,account_name,account_type,department_code,department,creator,owner_name,email,passphrase_rotate,status,remarks,created_at,updated_at,last_confirmed_at,confirm_token,token_expiry FROM privileged_accounts WHERE confirm_token=?`, token)
	var a models.PrivilegedAccount
	var lc, te sql.NullTime
	var tok sql.NullString
	if err := row.Scan(&a.ID, &a.SystemName, &a.Environment, &a.IPAddress, &a.InventoryDate, &a.AccountName, &a.AccountType, &a.DepartmentCode, &a.Department, &a.Creator, &a.OwnerName, &a.Email, &a.PassphraseRotate, &a.Status, &a.Remarks, &a.CreatedAt, &a.UpdatedAt, &lc, &tok, &te); err != nil {
		return nil, err
	}
	if lc.Valid {
		a.LastConfirmedAt = &lc.Time
	}
	if tok.Valid {
		a.ConfirmToken = tok.String
	}
	if te.Valid {
		a.TokenExpiry = &te.Time
	}
	return &a, nil
}

func (d *DB) CreateAccount(a *models.PrivilegedAccount) (int64, error) {
	res, err := d.conn.Exec(`INSERT INTO privileged_accounts (system_name,environment,ip_address,inventory_date,account_name,account_type,department_code,department,creator,owner_name,email,passphrase_rotate,status,remarks) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		a.SystemName, a.Environment, a.IPAddress, a.InventoryDate, a.AccountName, a.AccountType, a.DepartmentCode, a.Department, a.Creator, a.OwnerName, a.Email, a.PassphraseRotate, a.Status, a.Remarks)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) UpdateAccount(a *models.PrivilegedAccount) error {
	_, err := d.conn.Exec(`UPDATE privileged_accounts SET system_name=?,environment=?,ip_address=?,inventory_date=?,account_name=?,account_type=?,department_code=?,department=?,creator=?,owner_name=?,email=?,passphrase_rotate=?,status=?,remarks=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		a.SystemName, a.Environment, a.IPAddress, a.InventoryDate, a.AccountName, a.AccountType, a.DepartmentCode, a.Department, a.Creator, a.OwnerName, a.Email, a.PassphraseRotate, a.Status, a.Remarks, a.ID)
	return err
}

func (d *DB) DeleteAccount(id int) error {
	_, err := d.conn.Exec(`DELETE FROM privileged_accounts WHERE id=?`, id)
	return err
}

func (d *DB) SetConfirmToken(id int, token string, expiry time.Time) error {
	_, err := d.conn.Exec(`UPDATE privileged_accounts SET confirm_token=?,token_expiry=?,status='pending',updated_at=CURRENT_TIMESTAMP WHERE id=?`, token, expiry, id)
	return err
}

func (d *DB) ConfirmAccount(id int, action string) error {
	now := time.Now()
	newStatus := models.StatusActive
	if action == "stop" {
		newStatus = models.StatusClosed
	}
	_, err := d.conn.Exec(`UPDATE privileged_accounts SET last_confirmed_at=?,status=?,confirm_token=NULL,token_expiry=NULL,updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		now, newStatus, id)
	return err
}

// ---- Notification Logs ----

func (d *DB) AddNotificationLog(log *models.NotificationLog) error {
	_, err := d.conn.Exec(`INSERT INTO notification_logs (account_id,account_name,email,status,message) VALUES (?,?,?,?,?)`,
		log.AccountID, log.AccountName, log.Email, log.Status, log.Message)
	return err
}

func (d *DB) ListNotificationLogs() ([]models.NotificationLog, error) {
	rows, err := d.conn.Query(`SELECT id,account_id,account_name,email,sent_at,status,message FROM notification_logs ORDER BY sent_at DESC LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.NotificationLog
	for rows.Next() {
		var l models.NotificationLog
		rows.Scan(&l.ID, &l.AccountID, &l.AccountName, &l.Email, &l.SentAt, &l.Status, &l.Message)
		list = append(list, l)
	}
	return list, nil
}

// Stats 統計數字
func (d *DB) Stats() (map[string]int, error) {
	stats := map[string]int{}
	rows, err := d.conn.Query(`SELECT status, COUNT(*) FROM privileged_accounts GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	total := 0
	for rows.Next() {
		var s string
		var c int
		rows.Scan(&s, &c)
		stats[s] = c
		total += c
	}
	stats["total"] = total
	return stats, nil
}

func defaultFocusItems() models.DashboardFocusItems {
	return models.DashboardFocusItems{
		ActiveTitle:  "有效資料",
		ActiveMeta:   "顯示此表單目前有效或啟用中的資料筆數",
		PendingTitle: "待處理事項",
		PendingMeta:  "顯示此表單待辦、待確認或待補件數量",
		ClosedTitle:  "已完成事項",
		ClosedMeta:   "顯示此表單已完成、已結案或已停用數量",
		RecentTitle:  "最近更新紀錄",
	}
}

func (d *DB) seedDashboardForms() error {
	var count int
	if err := d.conn.QueryRow(`SELECT COUNT(*) FROM dashboard_forms`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	defaultForms := []models.DashboardForm{
		{
			Key:              "isms-04-062",
			Code:             "ISMS-04-062",
			ShortCode:        "04-062",
			Name:             "特殊權限帳號管理",
			Description:      "管理特殊權限帳號清冊、使用狀態、盤點日期與使用者確認進度。",
			DetailTitle:      "最近盤點清單",
			EmptyText:        "尚無特殊權限帳號資料",
			ProviderKey:      "privileged_accounts",
			DisplayOrder:     1,
			Enabled:          true,
			FocusItems: models.DashboardFocusItems{
				ActiveTitle:  "使用中帳號",
				ActiveMeta:   "目前仍在使用中的特殊權限帳號",
				PendingTitle: "待確認帳號",
				PendingMeta:  "建議優先通知與追蹤回覆",
				ClosedTitle:  "已關閉帳號",
				ClosedMeta:   "已完成停用或關閉流程",
				RecentTitle:  "最近更新紀錄",
			},
		},
		{
			Key:                      "isms-08-001-template",
			Code:                     "ISMS-08-001",
			ShortCode:                "08-001",
			Name:                     "示範表單骨架",
			Description:              "這是一張預留給後續擴充的示範表單卡，可替換為其他 ISMS 表單模組。",
			StatusNormalText:         "待建置",
			StatusNeedsAttentionText: "待建置",
			ProviderKey:              "placeholder",
			DisplayOrder:             2,
			Enabled:                  true,
			FocusItems: models.DashboardFocusItems{
				ActiveTitle:  "已建資料",
				ActiveMeta:   "未來可顯示此表單的有效資料筆數",
				PendingTitle: "待處理事項",
				PendingMeta:  "未來可顯示此表單的待辦或待確認數量",
				ClosedTitle:  "已完成事項",
				ClosedMeta:   "未來可顯示已完成或已結案數量",
				RecentTitle:  "建置狀態",
			},
		},
	}
	for _, form := range defaultForms {
		if _, err := d.CreateDashboardForm(&form); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) seedCustomTableTemplateRecords() error {
	var count int
	if err := d.conn.QueryRow(`SELECT COUNT(*) FROM custom_table_template_records`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := d.conn.Exec(`INSERT INTO custom_table_template_records (title,category,owner_name,status,inventory_date,email) VALUES
		('示範資料一', '範本類別', '王小明', 'active', '20260810', 'demo1@example.com'),
		('示範資料二', '範本類別', '李小華', 'pending', '20260809', 'demo2@example.com')`)
	return err
}

func marshalFocusItems(items models.DashboardFocusItems) string {
	if items == (models.DashboardFocusItems{}) {
		items = defaultFocusItems()
	}
	b, _ := json.Marshal(items)
	return string(b)
}

func unmarshalFocusItems(raw string) models.DashboardFocusItems {
	items := defaultFocusItems()
	if raw == "" {
		return items
	}
	_ = json.Unmarshal([]byte(raw), &items)
	return items
}

func scanDashboardForm(scanner interface{ Scan(dest ...interface{}) error }, f *models.DashboardForm) error {
	var enabled int
	var focusRaw string
	if err := scanner.Scan(
		&f.ID, &f.Key, &f.Code, &f.ShortCode, &f.Name, &f.Description,
		&f.DetailTitle, &f.EmptyText, &f.StatusNormalText, &f.StatusNeedsAttentionText,
		&f.ProviderKey, &f.DisplayOrder, &enabled, &focusRaw, &f.CreatedAt, &f.UpdatedAt,
	); err != nil {
		return err
	}
	f.Enabled = enabled == 1
	f.FocusItems = unmarshalFocusItems(focusRaw)
	return nil
}

func (d *DB) ListDashboardForms() ([]models.DashboardForm, error) {
	rows, err := d.conn.Query(`SELECT id,form_key,code,short_code,name,description,detail_title,empty_text,status_normal_text,status_needs_attention_text,provider_key,display_order,enabled,focus_items_json,created_at,updated_at FROM dashboard_forms ORDER BY display_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.DashboardForm
	for rows.Next() {
		var f models.DashboardForm
		if err := scanDashboardForm(rows, &f); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	if list == nil {
		list = []models.DashboardForm{}
	}
	return list, nil
}

func (d *DB) GetDashboardForm(id int) (*models.DashboardForm, error) {
	row := d.conn.QueryRow(`SELECT id,form_key,code,short_code,name,description,detail_title,empty_text,status_normal_text,status_needs_attention_text,provider_key,display_order,enabled,focus_items_json,created_at,updated_at FROM dashboard_forms WHERE id=?`, id)
	var f models.DashboardForm
	if err := scanDashboardForm(row, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

func (d *DB) CreateDashboardForm(f *models.DashboardForm) (int64, error) {
	res, err := d.conn.Exec(`INSERT INTO dashboard_forms (form_key,code,short_code,name,description,detail_title,empty_text,status_normal_text,status_needs_attention_text,provider_key,display_order,enabled,focus_items_json,updated_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP)`,
		f.Key, f.Code, f.ShortCode, f.Name, f.Description, f.DetailTitle, f.EmptyText, f.StatusNormalText, f.StatusNeedsAttentionText, f.ProviderKey, f.DisplayOrder, boolToInt(f.Enabled), marshalFocusItems(f.FocusItems))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) UpdateDashboardForm(f *models.DashboardForm) error {
	_, err := d.conn.Exec(`UPDATE dashboard_forms SET form_key=?,code=?,short_code=?,name=?,description=?,detail_title=?,empty_text=?,status_normal_text=?,status_needs_attention_text=?,provider_key=?,display_order=?,enabled=?,focus_items_json=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		f.Key, f.Code, f.ShortCode, f.Name, f.Description, f.DetailTitle, f.EmptyText, f.StatusNormalText, f.StatusNeedsAttentionText, f.ProviderKey, f.DisplayOrder, boolToInt(f.Enabled), marshalFocusItems(f.FocusItems), f.ID)
	return err
}

func (d *DB) DeleteDashboardForm(id int) error {
	_, err := d.conn.Exec(`DELETE FROM dashboard_forms WHERE id=?`, id)
	return err
}

func (d *DB) UpdateDashboardFormOrder(ids []int) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for idx, id := range ids {
		if _, err := tx.Exec(`UPDATE dashboard_forms SET display_order=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`, idx+1, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func (d *DB) Close() error {
	return d.conn.Close()
}

// SearchAccounts 搜尋
func (d *DB) SearchAccounts(keyword string) ([]models.PrivilegedAccount, error) {
	like := "%" + keyword + "%"
	rows, err := d.conn.Query(`SELECT id,system_name,environment,ip_address,inventory_date,account_name,account_type,department_code,department,creator,owner_name,email,passphrase_rotate,status,remarks,created_at,updated_at,last_confirmed_at FROM privileged_accounts WHERE account_name LIKE ? OR owner_name LIKE ? OR system_name LIKE ? OR department LIKE ? OR department_code LIKE ? ORDER BY id DESC`,
		like, like, like, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.PrivilegedAccount
	for rows.Next() {
		var a models.PrivilegedAccount
		var lc sql.NullTime
		rows.Scan(&a.ID, &a.SystemName, &a.Environment, &a.IPAddress, &a.InventoryDate, &a.AccountName, &a.AccountType, &a.DepartmentCode, &a.Department, &a.Creator, &a.OwnerName, &a.Email, &a.PassphraseRotate, &a.Status, &a.Remarks, &a.CreatedAt, &a.UpdatedAt, &lc)
		if lc.Valid {
			a.LastConfirmedAt = &lc.Time
		}
		list = append(list, a)
	}
	return list, nil
}

// BulkSetPending 批次設定 pending 狀態發送通知
func (d *DB) GetActiveAccountsWithEmail() ([]models.PrivilegedAccount, error) {
	rows, err := d.conn.Query(`SELECT id,system_name,environment,ip_address,inventory_date,account_name,account_type,department_code,department,creator,owner_name,email,passphrase_rotate,status,remarks,created_at,updated_at,last_confirmed_at FROM privileged_accounts WHERE status='active' AND email != '' ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.PrivilegedAccount
	for rows.Next() {
		var a models.PrivilegedAccount
		var lc sql.NullTime
		rows.Scan(&a.ID, &a.SystemName, &a.Environment, &a.IPAddress, &a.InventoryDate, &a.AccountName, &a.AccountType, &a.DepartmentCode, &a.Department, &a.Creator, &a.OwnerName, &a.Email, &a.PassphraseRotate, &a.Status, &a.Remarks, &a.CreatedAt, &a.UpdatedAt, &lc)
		if lc.Valid {
			a.LastConfirmedAt = &lc.Time
		}
		list = append(list, a)
	}
	return list, nil
}

// PendingAccounts 取得過期待確認帳號
func (d *DB) GetExpiredPendingAccounts() ([]models.PrivilegedAccount, error) {
	rows, err := d.conn.Query(`SELECT id,system_name,account_name,owner_name,email,status FROM privileged_accounts WHERE status='pending' AND token_expiry < CURRENT_TIMESTAMP`)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()
	var list []models.PrivilegedAccount
	for rows.Next() {
		var a models.PrivilegedAccount
		rows.Scan(&a.ID, &a.SystemName, &a.AccountName, &a.OwnerName, &a.Email, &a.Status)
		list = append(list, a)
	}
	return list, nil
}

// ListCustomTableTemplateRecords 自訂資料表 provider 範本查詢介面
//
// Provider 開發指南：
// 1. 未來新增真實 provider 時，可複製這個函式並改成 ListXxxRecords。
// 2. 查詢結果請回傳該 provider 專屬的資料模型 slice，不要直接回傳 DashboardRecord。
// 3. 資料表到 DashboardRecord 的映射統一放在 handlers.go，避免 db 層摻雜首頁展示邏輯。
// 4. 若需要額外欄位，優先加在 provider 自己的 struct，再由 handler 決定要不要映射到首頁。
func (d *DB) ListCustomTableTemplateRecords() ([]models.CustomTableTemplateRecord, error) {
	rows, err := d.conn.Query(`SELECT id,title,category,owner_name,status,inventory_date,email,created_at,updated_at FROM custom_table_template_records ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.CustomTableTemplateRecord
	for rows.Next() {
		var r models.CustomTableTemplateRecord
		if err := rows.Scan(&r.ID, &r.Title, &r.Category, &r.OwnerName, &r.Status, &r.InventoryDate, &r.Email, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	if list == nil {
		list = []models.CustomTableTemplateRecord{}
	}
	return list, nil
}
