package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"isms-privilege/internal/models"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

type creatorBackfillTarget struct {
	table string
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

	CREATE TABLE IF NOT EXISTS operation_logs (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		event_type     TEXT    NOT NULL DEFAULT '',
		occurred_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		location       TEXT    NOT NULL DEFAULT '',
		request_path   TEXT    NOT NULL DEFAULT '',
		request_method TEXT    NOT NULL DEFAULT '',
		status_code    INTEGER NOT NULL DEFAULT 200,
		source_ip      TEXT    NOT NULL DEFAULT '',
		user_agent     TEXT    NOT NULL DEFAULT '',
		user_email     TEXT    NOT NULL DEFAULT '',
		user_name      TEXT    NOT NULL DEFAULT '',
		user_google_id TEXT    NOT NULL DEFAULT '',
		department     TEXT    NOT NULL DEFAULT ''
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

	CREATE TABLE IF NOT EXISTS system_platform_requests (
		id                     INTEGER PRIMARY KEY AUTOINCREMENT,
		request_date           TEXT    NOT NULL DEFAULT '',
		applicant_name         TEXT    NOT NULL DEFAULT '',
		applicant_department   TEXT    NOT NULL DEFAULT '',
		applicant_title        TEXT    NOT NULL DEFAULT '',
		office_phone           TEXT    NOT NULL DEFAULT '',
		email                  TEXT    NOT NULL DEFAULT '',
		pi_name                TEXT    NOT NULL DEFAULT '',
		system_name            TEXT    NOT NULL DEFAULT '',
		system_alias           TEXT    NOT NULL DEFAULT '',
		system_purpose         TEXT    NOT NULL DEFAULT '',
		estimated_users        TEXT    NOT NULL DEFAULT '',
		internal_only          TEXT    NOT NULL DEFAULT '是',
		ip_restriction         TEXT    NOT NULL DEFAULT '',
		request_start_date     TEXT    NOT NULL DEFAULT '',
		request_end_date       TEXT    NOT NULL DEFAULT '',
		request_type           TEXT    NOT NULL DEFAULT '上架新增',
		shutdown_retain_months TEXT    NOT NULL DEFAULT '',
		shutdown_reason        TEXT    NOT NULL DEFAULT '',
		environment_type       TEXT    NOT NULL DEFAULT '正式環境',
		operating_system       TEXT    NOT NULL DEFAULT 'Rocky 9',
		operating_system_other TEXT    NOT NULL DEFAULT '',
		disk_size              TEXT    NOT NULL DEFAULT '',
		special_requirements   TEXT    NOT NULL DEFAULT '',
		domain_settings        TEXT    NOT NULL DEFAULT '',
		other_requirements     TEXT    NOT NULL DEFAULT '',
		backup_required        TEXT    NOT NULL DEFAULT '是',
		backup_requirements    TEXT    NOT NULL DEFAULT '',
		backup_reason          TEXT    NOT NULL DEFAULT '',
		applicant_signature    TEXT    NOT NULL DEFAULT '',
		supervisor_signature   TEXT    NOT NULL DEFAULT '',
		status                 TEXT    NOT NULL DEFAULT 'active',
		creator                TEXT    NOT NULL DEFAULT '',
		remarks                TEXT    NOT NULL DEFAULT '',
		created_at             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at             DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS application_change_requests (
		id                            INTEGER PRIMARY KEY AUTOINCREMENT,
		suggestor                     TEXT    NOT NULL DEFAULT '',
		form_date                     TEXT    NOT NULL DEFAULT '',
		approver                      TEXT    NOT NULL DEFAULT '',
		system_name                   TEXT    NOT NULL DEFAULT '',
		feature_name                  TEXT    NOT NULL DEFAULT '',
		is_required_feature           TEXT    NOT NULL DEFAULT '',
		is_major_impact               TEXT    NOT NULL DEFAULT '',
		expected_online_date          TEXT    NOT NULL DEFAULT '',
		background_description        TEXT    NOT NULL DEFAULT '',
		existing_security_measures    TEXT    NOT NULL DEFAULT '',
		security_scope_involved       TEXT    NOT NULL DEFAULT '',
		access_control_measures       TEXT    NOT NULL DEFAULT '',
		audit_measures                TEXT    NOT NULL DEFAULT '',
		continuity_measures           TEXT    NOT NULL DEFAULT '',
		identification_measures       TEXT    NOT NULL DEFAULT '',
		system_acquisition_measures   TEXT    NOT NULL DEFAULT '',
		communication_measures        TEXT    NOT NULL DEFAULT '',
		integrity_measures            TEXT    NOT NULL DEFAULT '',
		capacity_management           TEXT    NOT NULL DEFAULT '',
		capacity_change_description   TEXT    NOT NULL DEFAULT '',
		other_security_measures       TEXT    NOT NULL DEFAULT '',
		other_security_description    TEXT    NOT NULL DEFAULT '',
		information_service_opinion   TEXT    NOT NULL DEFAULT '',
		meeting_time                  TEXT    NOT NULL DEFAULT '',
		meeting_decision              TEXT    NOT NULL DEFAULT '',
		rejection_reason              TEXT    NOT NULL DEFAULT '',
		coordinating_staff            TEXT    NOT NULL DEFAULT '',
		coordination_date             TEXT    NOT NULL DEFAULT '',
		coordination_approver         TEXT    NOT NULL DEFAULT '',
		status                        TEXT    NOT NULL DEFAULT 'active',
		creator                       TEXT    NOT NULL DEFAULT '',
		remarks                       TEXT    NOT NULL DEFAULT '',
		created_at                    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at                    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS firewall_requests (
		id                   INTEGER PRIMARY KEY AUTOINCREMENT,
		legacy_form_number   TEXT    NOT NULL DEFAULT '',
		system_name          TEXT    NOT NULL DEFAULT '',
		action               TEXT    NOT NULL DEFAULT '',
		purpose_type         TEXT    NOT NULL DEFAULT '',
		source_zone          TEXT    NOT NULL DEFAULT '',
		source_zone2         TEXT    NOT NULL DEFAULT '',
		source_ip            TEXT    NOT NULL DEFAULT '',
		destination_zone     TEXT    NOT NULL DEFAULT '',
		destination_zone2    TEXT    NOT NULL DEFAULT '',
		destination_ip       TEXT    NOT NULL DEFAULT '',
		protocol_type        TEXT    NOT NULL DEFAULT '',
		start_date           TEXT    NOT NULL DEFAULT '',
		end_date             TEXT    NOT NULL DEFAULT '',
		request_date         TEXT    NOT NULL DEFAULT '',
		rule_description     TEXT    NOT NULL DEFAULT '',
		firewall_zone        TEXT    NOT NULL DEFAULT '',
		firewall_id          TEXT    NOT NULL DEFAULT '',
		status               TEXT    NOT NULL DEFAULT 'active',
		creator              TEXT    NOT NULL DEFAULT '',
		remarks              TEXT    NOT NULL DEFAULT '',
		created_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS asset_inventory_records (
		id                            INTEGER PRIMARY KEY AUTOINCREMENT,
		system_name                   TEXT    NOT NULL DEFAULT '',
		environment                   TEXT    NOT NULL DEFAULT '正式',
		asset_code                    TEXT    NOT NULL DEFAULT '',
		asset_type                    TEXT    NOT NULL DEFAULT '',
		asset_name                    TEXT    NOT NULL DEFAULT '',
		vendor_name                   TEXT    NOT NULL DEFAULT '',
		is_core_asset                 TEXT    NOT NULL DEFAULT '否',
		has_national_security_concern TEXT    NOT NULL DEFAULT '否',
		asset_description             TEXT    NOT NULL DEFAULT '',
		quantity                      TEXT    NOT NULL DEFAULT '1',
		os_config_baseline            TEXT    NOT NULL DEFAULT '',
		browser_config_baseline       TEXT    NOT NULL DEFAULT '',
		network_config_baseline       TEXT    NOT NULL DEFAULT '',
		application_config_baseline   TEXT    NOT NULL DEFAULT '',
		other_config_baseline         TEXT    NOT NULL DEFAULT '',
		config_exception_code         TEXT    NOT NULL DEFAULT '',
		manager_department            TEXT    NOT NULL DEFAULT '',
		user_department               TEXT    NOT NULL DEFAULT '',
		location                      TEXT    NOT NULL DEFAULT '',
		confidentiality               TEXT    NOT NULL DEFAULT '',
		integrity                     TEXT    NOT NULL DEFAULT '',
		availability                  TEXT    NOT NULL DEFAULT '',
		asset_value                   TEXT    NOT NULL DEFAULT '',
		legal_compliance              TEXT    NOT NULL DEFAULT '',
		protection_level              TEXT    NOT NULL DEFAULT '',
		mtpd                          TEXT    NOT NULL DEFAULT '',
		rto                           TEXT    NOT NULL DEFAULT '',
		rpo                           TEXT    NOT NULL DEFAULT '',
		status                        TEXT    NOT NULL DEFAULT 'active',
		creator                       TEXT    NOT NULL DEFAULT '',
		remarks                       TEXT    NOT NULL DEFAULT '',
		created_at                    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at                    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS protection_baseline_records (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		system_name    TEXT    NOT NULL DEFAULT '',
		security_level TEXT    NOT NULL DEFAULT '普',
		filled_by      TEXT    NOT NULL DEFAULT '',
		form_date      TEXT    NOT NULL DEFAULT '',
		reviewer       TEXT    NOT NULL DEFAULT '',
		review_date    TEXT    NOT NULL DEFAULT '',
		status         TEXT    NOT NULL DEFAULT 'active',
		creator        TEXT    NOT NULL DEFAULT '',
		remarks        TEXT    NOT NULL DEFAULT '',
		created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS protection_baseline_controls (
		id                   INTEGER PRIMARY KEY AUTOINCREMENT,
		record_id            INTEGER NOT NULL REFERENCES protection_baseline_records(id) ON DELETE CASCADE,
		item_no              INTEGER NOT NULL,
		domain_name          TEXT    NOT NULL DEFAULT '',
		control_category     TEXT    NOT NULL DEFAULT '',
		requirement_level    TEXT    NOT NULL DEFAULT '',
		control_description  TEXT    NOT NULL DEFAULT '',
		measure_notes        TEXT    NOT NULL DEFAULT '',
		applies              TEXT    NOT NULL DEFAULT '',
		implementation_notes TEXT    NOT NULL DEFAULT '',
		compliance           TEXT    NOT NULL DEFAULT '',
		finding              TEXT    NOT NULL DEFAULT '',
		remarks              TEXT    NOT NULL DEFAULT ''
	);
	`
	_, err := d.conn.Exec(schema)
	_, _ = d.conn.Exec(`ALTER TABLE privileged_accounts ADD COLUMN environment TEXT NOT NULL DEFAULT '正式區'`)
	_, _ = d.conn.Exec(`ALTER TABLE privileged_accounts ADD COLUMN department_code TEXT NOT NULL DEFAULT ''`)
	_, _ = d.conn.Exec(`ALTER TABLE privileged_accounts ADD COLUMN creator TEXT NOT NULL DEFAULT ''`)
	_, _ = d.conn.Exec(`ALTER TABLE asset_inventory_records ADD COLUMN environment TEXT NOT NULL DEFAULT '正式'`)
	if err == nil {
		_ = d.seedDashboardForms()
		_ = d.seedCustomTableTemplateRecords()
		_ = d.seedFirewallRequests()
		_ = d.seedAssetInventoryRecords()
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
	return d.ListAccountsByCreator(status, "")
}

func (d *DB) ListAccountsByCreator(status, creator string) ([]models.PrivilegedAccount, error) {
	q := `SELECT id,system_name,environment,ip_address,inventory_date,account_name,account_type,department_code,department,creator,owner_name,email,passphrase_rotate,status,remarks,created_at,updated_at,last_confirmed_at FROM privileged_accounts`
	args := []interface{}{}
	clauses := []string{}
	if status != "" && status != "all" {
		clauses = append(clauses, "status = ?")
		args = append(args, status)
	}
	if strings.TrimSpace(creator) != "" {
		clauses = append(clauses, "creator = ?")
		args = append(args, creator)
	}
	if len(clauses) > 0 {
		q += " WHERE " + strings.Join(clauses, " AND ")
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
	return d.GetAccountByCreator(id, "")
}

func (d *DB) GetAccountByCreator(id int, creator string) (*models.PrivilegedAccount, error) {
	query := `SELECT id,system_name,environment,ip_address,inventory_date,account_name,account_type,department_code,department,creator,owner_name,email,passphrase_rotate,status,remarks,created_at,updated_at,last_confirmed_at,confirm_token,token_expiry FROM privileged_accounts WHERE id=?`
	args := []interface{}{id}
	if strings.TrimSpace(creator) != "" {
		query += ` AND creator=?`
		args = append(args, creator)
	}
	row := d.conn.QueryRow(query, args...)
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
	return d.DeleteAccountByCreator(id, "")
}

func (d *DB) DeleteAccountByCreator(id int, creator string) error {
	query := `DELETE FROM privileged_accounts WHERE id=?`
	args := []interface{}{id}
	if strings.TrimSpace(creator) != "" {
		query += ` AND creator=?`
		args = append(args, creator)
	}
	_, err := d.conn.Exec(query, args...)
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

func (d *DB) AddOperationLog(log *models.OperationLog) error {
	_, err := d.conn.Exec(`INSERT INTO operation_logs (event_type,location,request_path,request_method,status_code,source_ip,user_agent,user_email,user_name,user_google_id,department) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		log.EventType, log.Location, log.RequestPath, log.RequestMethod, log.StatusCode, log.SourceIP, log.UserAgent, log.UserEmail, log.UserName, log.UserGoogleID, log.Department)
	return err
}

func (d *DB) ListOperationLogs(limit int) ([]models.OperationLog, error) {
	return d.ListOperationLogsByUser(limit, "")
}

func (d *DB) ListOperationLogsByUser(limit int, userEmail string) ([]models.OperationLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	query := `SELECT id,event_type,occurred_at,location,request_path,request_method,status_code,source_ip,user_agent,user_email,user_name,user_google_id,department FROM operation_logs`
	args := []interface{}{}
	if strings.TrimSpace(userEmail) != "" {
		query += ` WHERE user_email = ?`
		args = append(args, userEmail)
	}
	query += ` ORDER BY occurred_at DESC, id DESC LIMIT ?`
	args = append(args, limit)
	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.OperationLog
	for rows.Next() {
		var item models.OperationLog
		if err := rows.Scan(&item.ID, &item.EventType, &item.OccurredAt, &item.Location, &item.RequestPath, &item.RequestMethod, &item.StatusCode, &item.SourceIP, &item.UserAgent, &item.UserEmail, &item.UserName, &item.UserGoogleID, &item.Department); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	if list == nil {
		list = []models.OperationLog{}
	}
	return list, nil
}

func (d *DB) ListNotificationLogs() ([]models.NotificationLog, error) {
	return d.ListNotificationLogsByCreator("")
}

func (d *DB) ListNotificationLogsByCreator(creator string) ([]models.NotificationLog, error) {
	query := `SELECT l.id,l.account_id,l.account_name,l.email,l.sent_at,l.status,l.message
		FROM notification_logs l
		JOIN privileged_accounts a ON a.id = l.account_id`
	args := []interface{}{}
	if strings.TrimSpace(creator) != "" {
		query += ` WHERE a.creator = ?`
		args = append(args, creator)
	}
	query += ` ORDER BY l.sent_at DESC LIMIT 200`
	rows, err := d.conn.Query(query, args...)
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
	return d.StatsByCreator("")
}

func (d *DB) StatsByCreator(creator string) (map[string]int, error) {
	stats := map[string]int{}
	query := `SELECT status, COUNT(*) FROM privileged_accounts`
	args := []interface{}{}
	if strings.TrimSpace(creator) != "" {
		query += ` WHERE creator = ?`
		args = append(args, creator)
	}
	query += ` GROUP BY status`
	rows, err := d.conn.Query(query, args...)
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
	defaultForms := []models.DashboardForm{
		{
			Key:          "isms-04-008",
			Code:         "ISMS-04-008",
			ShortCode:    "04-008",
			Name:         "資訊資產清冊",
			Description:  "管理資訊資產主檔、資產價值、組態基準與營運關鍵指標。",
			DetailTitle:  "最近資產資料",
			EmptyText:    "尚無資訊資產清冊資料",
			ProviderKey:  "asset_inventory",
			DisplayOrder: 1,
			Enabled:      true,
			FocusItems: models.DashboardFocusItems{
				ActiveTitle:  "在管資產",
				ActiveMeta:   "目前仍納入清冊管理的資訊資產",
				PendingTitle: "待補資料",
				PendingMeta:  "尚待確認、補件或盤點中的資產",
				ClosedTitle:  "已下架資產",
				ClosedMeta:   "已停用、淘汰或結案的資產",
				RecentTitle:  "最近更新資產",
			},
		},
		{
			Key:          "isms-04-062",
			Code:         "ISMS-04-062",
			ShortCode:    "04-062",
			Name:         "特殊權限帳號管理",
			Description:  "管理特殊權限帳號清冊、使用狀態、盤點日期與使用者確認進度。",
			DetailTitle:  "最近盤點清單",
			EmptyText:    "尚無特殊權限帳號資料",
			ProviderKey:  "privileged_accounts",
			DisplayOrder: 2,
			Enabled:      true,
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
			Key:          "isms-04-042",
			Code:         "ISMS-04-042",
			ShortCode:    "04-042",
			Name:         "防火牆申請單",
			Description:  "管理防火牆規則申請、來源與目的區域、IP、通訊埠與有效期間。",
			DetailTitle:  "最近防火牆申請",
			EmptyText:    "尚無防火牆申請資料",
			ProviderKey:  "firewall_requests",
			DisplayOrder: 3,
			Enabled:      true,
			FocusItems: models.DashboardFocusItems{
				ActiveTitle:  "生效中規則",
				ActiveMeta:   "目前正在使用或生效中的防火牆申請",
				PendingTitle: "待追蹤申請",
				PendingMeta:  "待審核、待補件或待確認的規則申請",
				ClosedTitle:  "已完成申請",
				ClosedMeta:   "已結案、停用或到期後完成處理的申請",
				RecentTitle:  "最近申請紀錄",
			},
		},
		{
			Key:          "isms-04-069",
			Code:         "ISMS-04-069",
			ShortCode:    "04-069",
			Name:         "資通系統防護基準執行說明表",
			Description:  "管理系統安全等級與 80 項資通系統防護基準控制措施執行情形。",
			DetailTitle:  "最近防護基準資料",
			EmptyText:    "尚無 04-069 防護基準資料",
			ProviderKey:  "protection_baselines",
			DisplayOrder: 4,
			Enabled:      true,
			FocusItems: models.DashboardFocusItems{
				ActiveTitle:  "進行中表單",
				ActiveMeta:   "目前仍持續維護與評估中的 04-069 表單",
				PendingTitle: "待補強事項",
				PendingMeta:  "需補件、改善或待完成填寫的 04-069 表單",
				ClosedTitle:  "已完成表單",
				ClosedMeta:   "已完成審查或結案的 04-069 表單",
				RecentTitle:  "最近更新表單",
			},
		},
		{
			Key:          "isms-04-078",
			Code:         "ISMS-04-078",
			ShortCode:    "04-078",
			Name:         "系統平台申請",
			Description:  "記錄主機申請、上下架、環境、OS、備份與特殊需求等資料。",
			DetailTitle:  "最近申請資料",
			EmptyText:    "尚無系統平台申請資料",
			ProviderKey:  "system_platform_requests",
			DisplayOrder: 5,
			Enabled:      true,
			FocusItems: models.DashboardFocusItems{
				ActiveTitle:  "進行中申請",
				ActiveMeta:   "目前正在處理或使用中的平台申請",
				PendingTitle: "待追蹤申請",
				PendingMeta:  "需補件、待審核或待確認的申請",
				ClosedTitle:  "已完成申請",
				ClosedMeta:   "已完成下架、關機或結案的申請",
				RecentTitle:  "最近申請紀錄",
			},
		},
		{
			Key:          "isms-04-052",
			Code:         "ISMS-04-052",
			ShortCode:    "04-052",
			Name:         "應用系統功能需求更新建議單",
			Description:  "記錄功能需求屬性、問題背景、資安措施評估與資訊服務處意見。",
			DetailTitle:  "最近需求建議",
			EmptyText:    "尚無應用系統功能需求更新建議資料",
			ProviderKey:  "application_change_requests",
			DisplayOrder: 6,
			Enabled:      true,
			FocusItems: models.DashboardFocusItems{
				ActiveTitle:  "進行中建議",
				ActiveMeta:   "目前正在評估或辦理中的功能需求建議",
				PendingTitle: "待商議建議",
				PendingMeta:  "需共同商議、補件或追蹤的需求建議",
				ClosedTitle:  "已完成建議",
				ClosedMeta:   "已同意辦理、完成或結案的需求建議",
				RecentTitle:  "最近建議紀錄",
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
			DisplayOrder:             7,
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
		var existingID int
		err := d.conn.QueryRow(`SELECT id FROM dashboard_forms WHERE form_key=?`, form.Key).Scan(&existingID)
		if err == nil {
			continue
		}
		if err != sql.ErrNoRows {
			return err
		}
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

func (d *DB) seedFirewallRequests() error {
	var count int
	if err := d.conn.QueryRow(`SELECT COUNT(*) FROM firewall_requests`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := d.conn.Exec(`INSERT INTO firewall_requests (
		legacy_form_number,system_name,action,purpose_type,source_zone,source_zone2,source_ip,
		destination_zone,destination_zone2,destination_ip,protocol_type,start_date,end_date,
		request_date,rule_description,firewall_zone,firewall_id,status
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		"", "AI PRO 管理系統", "開通", "常態性服務", "campus", "", "10.109.193.19/32",
		"資料中心", "I", "10.109.233.61/32", "TCP 80,443,22", "2026/06/16", "2027/09/15",
		"2026/06/10", "開通桌機至 AI PRO 管理系統主機連線", "機房:DC", "D-260610-1-1", "active")
	return err
}

func (d *DB) seedAssetInventoryRecords() error {
	var count int
	if err := d.conn.QueryRow(`SELECT COUNT(*) FROM asset_inventory_records`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := d.conn.Exec(`INSERT INTO asset_inventory_records (
		system_name,environment,asset_code,asset_type,asset_name,vendor_name,is_core_asset,has_national_security_concern,
		asset_description,quantity,manager_department,user_department,location,confidentiality,integrity,
		availability,asset_value,protection_level,status
	) VALUES
		('辦公用桌上型電腦','正式','ISMS-HW-F08','實體類','桌上型電腦','ASUS','否','否','否','1','推廣科/劉智漢','推廣科/劉智漢','行政大樓4102室','2','2','2','2','0','active'),
		('全院個人學習時數管理系統API（開發區）','開發','ISMS-SW-F02','軟體類','全院個人學習時數管理系統API（開發區）','','否','否','否','1','推廣科/劉智漢','全院及e等公務員平台','資訊服務處機房','1','1','1','1','0','active'),
		('全院個人學習時數管理系統（正式區）','正式','ISMS-SW-F03','軟體類','全院個人學習時數管理系統','','否','否','否','1','推廣科/劉智漢','全院','資訊服務處機房','1','1','1','1','0','active')`)
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

func scanDashboardForm(scanner interface {
	Scan(dest ...interface{}) error
}, f *models.DashboardForm) error {
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

func (d *DB) BackfillEmptyCreators(userEmail string) (*models.CreatorBackfillResult, error) {
	userEmail = strings.TrimSpace(userEmail)
	if userEmail == "" {
		return nil, fmt.Errorf("user email is required")
	}

	targets := []creatorBackfillTarget{
		{table: "privileged_accounts"},
		{table: "system_platform_requests"},
		{table: "firewall_requests"},
		{table: "asset_inventory_records"},
		{table: "protection_baseline_records"},
	}

	tx, err := d.conn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result := &models.CreatorBackfillResult{}
	for _, target := range targets {
		res, err := tx.Exec(fmt.Sprintf(`UPDATE %s SET creator=?, updated_at=CURRENT_TIMESTAMP WHERE TRIM(COALESCE(creator, ''))=''`, target.table), userEmail)
		if err != nil {
			return nil, err
		}
		affected64, err := res.RowsAffected()
		if err != nil {
			return nil, err
		}
		affected := int(affected64)
		result.TotalUpdated += affected
		switch target.table {
		case "privileged_accounts":
			result.PrivilegedAccounts = affected
		case "system_platform_requests":
			result.SystemPlatforms = affected
		case "firewall_requests":
			result.FirewallRequests = affected
		case "asset_inventory_records":
			result.AssetInventory = affected
		case "protection_baseline_records":
			result.ProtectionBaselines = affected
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return result, nil
}

// SearchAccounts 搜尋
func (d *DB) SearchAccounts(keyword string) ([]models.PrivilegedAccount, error) {
	return d.SearchAccountsByCreator(keyword, "")
}

func (d *DB) SearchAccountsByCreator(keyword, creator string) ([]models.PrivilegedAccount, error) {
	like := "%" + keyword + "%"
	query := `SELECT id,system_name,environment,ip_address,inventory_date,account_name,account_type,department_code,department,creator,owner_name,email,passphrase_rotate,status,remarks,created_at,updated_at,last_confirmed_at FROM privileged_accounts WHERE (account_name LIKE ? OR owner_name LIKE ? OR system_name LIKE ? OR department LIKE ? OR department_code LIKE ?)`
	args := []interface{}{like, like, like, like, like}
	if strings.TrimSpace(creator) != "" {
		query += ` AND creator = ?`
		args = append(args, creator)
	}
	query += ` ORDER BY id DESC`
	rows, err := d.conn.Query(query, args...)
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
	return d.GetActiveAccountsWithEmailByCreator("")
}

func (d *DB) GetActiveAccountsWithEmailByCreator(creator string) ([]models.PrivilegedAccount, error) {
	query := `SELECT id,system_name,environment,ip_address,inventory_date,account_name,account_type,department_code,department,creator,owner_name,email,passphrase_rotate,status,remarks,created_at,updated_at,last_confirmed_at FROM privileged_accounts WHERE status='active' AND email != ''`
	args := []interface{}{}
	if strings.TrimSpace(creator) != "" {
		query += ` AND creator = ?`
		args = append(args, creator)
	}
	query += ` ORDER BY id`
	rows, err := d.conn.Query(query, args...)
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

func (d *DB) ListSystemPlatformRequests() ([]models.SystemPlatformRequest, error) {
	return d.ListSystemPlatformRequestsByCreator("")
}

func (d *DB) ListSystemPlatformRequestsByCreator(creator string) ([]models.SystemPlatformRequest, error) {
	query := `SELECT id,request_date,applicant_name,applicant_department,applicant_title,office_phone,email,pi_name,system_name,system_alias,system_purpose,estimated_users,internal_only,ip_restriction,request_start_date,request_end_date,request_type,shutdown_retain_months,shutdown_reason,environment_type,operating_system,operating_system_other,disk_size,special_requirements,domain_settings,other_requirements,backup_required,backup_requirements,backup_reason,applicant_signature,supervisor_signature,status,creator,remarks,created_at,updated_at FROM system_platform_requests`
	args := []interface{}{}
	if strings.TrimSpace(creator) != "" {
		query += ` WHERE creator=?`
		args = append(args, creator)
	}
	query += ` ORDER BY id DESC`
	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.SystemPlatformRequest
	for rows.Next() {
		var r models.SystemPlatformRequest
		if err := rows.Scan(&r.ID, &r.RequestDate, &r.ApplicantName, &r.ApplicantDepartment, &r.ApplicantTitle, &r.OfficePhone, &r.Email, &r.PIName, &r.SystemName, &r.SystemAlias, &r.SystemPurpose, &r.EstimatedUsers, &r.InternalOnly, &r.IPRestriction, &r.RequestStartDate, &r.RequestEndDate, &r.RequestType, &r.ShutdownRetainMonths, &r.ShutdownReason, &r.EnvironmentType, &r.OperatingSystem, &r.OperatingSystemOther, &r.DiskSize, &r.SpecialRequirements, &r.DomainSettings, &r.OtherRequirements, &r.BackupRequired, &r.BackupRequirements, &r.BackupReason, &r.ApplicantSignature, &r.SupervisorSignature, &r.Status, &r.Creator, &r.Remarks, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	if list == nil {
		list = []models.SystemPlatformRequest{}
	}
	return list, nil
}

func (d *DB) GetSystemPlatformRequest(id int) (*models.SystemPlatformRequest, error) {
	return d.GetSystemPlatformRequestByCreator(id, "")
}

func (d *DB) GetSystemPlatformRequestByCreator(id int, creator string) (*models.SystemPlatformRequest, error) {
	query := `SELECT id,request_date,applicant_name,applicant_department,applicant_title,office_phone,email,pi_name,system_name,system_alias,system_purpose,estimated_users,internal_only,ip_restriction,request_start_date,request_end_date,request_type,shutdown_retain_months,shutdown_reason,environment_type,operating_system,operating_system_other,disk_size,special_requirements,domain_settings,other_requirements,backup_required,backup_requirements,backup_reason,applicant_signature,supervisor_signature,status,creator,remarks,created_at,updated_at FROM system_platform_requests WHERE id=?`
	args := []interface{}{id}
	if strings.TrimSpace(creator) != "" {
		query += ` AND creator=?`
		args = append(args, creator)
	}
	row := d.conn.QueryRow(query, args...)
	var r models.SystemPlatformRequest
	if err := row.Scan(&r.ID, &r.RequestDate, &r.ApplicantName, &r.ApplicantDepartment, &r.ApplicantTitle, &r.OfficePhone, &r.Email, &r.PIName, &r.SystemName, &r.SystemAlias, &r.SystemPurpose, &r.EstimatedUsers, &r.InternalOnly, &r.IPRestriction, &r.RequestStartDate, &r.RequestEndDate, &r.RequestType, &r.ShutdownRetainMonths, &r.ShutdownReason, &r.EnvironmentType, &r.OperatingSystem, &r.OperatingSystemOther, &r.DiskSize, &r.SpecialRequirements, &r.DomainSettings, &r.OtherRequirements, &r.BackupRequired, &r.BackupRequirements, &r.BackupReason, &r.ApplicantSignature, &r.SupervisorSignature, &r.Status, &r.Creator, &r.Remarks, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return nil, err
	}
	return &r, nil
}

func (d *DB) CreateSystemPlatformRequest(r *models.SystemPlatformRequest) (int64, error) {
	res, err := d.conn.Exec(`INSERT INTO system_platform_requests (request_date,applicant_name,applicant_department,applicant_title,office_phone,email,pi_name,system_name,system_alias,system_purpose,estimated_users,internal_only,ip_restriction,request_start_date,request_end_date,request_type,shutdown_retain_months,shutdown_reason,environment_type,operating_system,operating_system_other,disk_size,special_requirements,domain_settings,other_requirements,backup_required,backup_requirements,backup_reason,applicant_signature,supervisor_signature,status,creator,remarks) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		r.RequestDate, r.ApplicantName, r.ApplicantDepartment, r.ApplicantTitle, r.OfficePhone, r.Email, r.PIName, r.SystemName, r.SystemAlias, r.SystemPurpose, r.EstimatedUsers, r.InternalOnly, r.IPRestriction, r.RequestStartDate, r.RequestEndDate, r.RequestType, r.ShutdownRetainMonths, r.ShutdownReason, r.EnvironmentType, r.OperatingSystem, r.OperatingSystemOther, r.DiskSize, r.SpecialRequirements, r.DomainSettings, r.OtherRequirements, r.BackupRequired, r.BackupRequirements, r.BackupReason, r.ApplicantSignature, r.SupervisorSignature, r.Status, r.Creator, r.Remarks)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) UpdateSystemPlatformRequest(r *models.SystemPlatformRequest) error {
	_, err := d.conn.Exec(`UPDATE system_platform_requests SET request_date=?,applicant_name=?,applicant_department=?,applicant_title=?,office_phone=?,email=?,pi_name=?,system_name=?,system_alias=?,system_purpose=?,estimated_users=?,internal_only=?,ip_restriction=?,request_start_date=?,request_end_date=?,request_type=?,shutdown_retain_months=?,shutdown_reason=?,environment_type=?,operating_system=?,operating_system_other=?,disk_size=?,special_requirements=?,domain_settings=?,other_requirements=?,backup_required=?,backup_requirements=?,backup_reason=?,applicant_signature=?,supervisor_signature=?,status=?,creator=?,remarks=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		r.RequestDate, r.ApplicantName, r.ApplicantDepartment, r.ApplicantTitle, r.OfficePhone, r.Email, r.PIName, r.SystemName, r.SystemAlias, r.SystemPurpose, r.EstimatedUsers, r.InternalOnly, r.IPRestriction, r.RequestStartDate, r.RequestEndDate, r.RequestType, r.ShutdownRetainMonths, r.ShutdownReason, r.EnvironmentType, r.OperatingSystem, r.OperatingSystemOther, r.DiskSize, r.SpecialRequirements, r.DomainSettings, r.OtherRequirements, r.BackupRequired, r.BackupRequirements, r.BackupReason, r.ApplicantSignature, r.SupervisorSignature, r.Status, r.Creator, r.Remarks, r.ID)
	return err
}

func (d *DB) DeleteSystemPlatformRequest(id int) error {
	return d.DeleteSystemPlatformRequestByCreator(id, "")
}

func (d *DB) DeleteSystemPlatformRequestByCreator(id int, creator string) error {
	query := `DELETE FROM system_platform_requests WHERE id=?`
	args := []interface{}{id}
	if strings.TrimSpace(creator) != "" {
		query += ` AND creator=?`
		args = append(args, creator)
	}
	_, err := d.conn.Exec(query, args...)
	return err
}

func (d *DB) ListFirewallRequests() ([]models.FirewallRequest, error) {
	return d.ListFirewallRequestsByCreator("")
}

func (d *DB) ListFirewallRequestsByCreator(creator string) ([]models.FirewallRequest, error) {
	query := `SELECT id,legacy_form_number,system_name,action,purpose_type,source_zone,source_zone2,source_ip,destination_zone,destination_zone2,destination_ip,protocol_type,start_date,end_date,request_date,rule_description,firewall_zone,firewall_id,status,creator,remarks,created_at,updated_at FROM firewall_requests`
	args := []interface{}{}
	if strings.TrimSpace(creator) != "" {
		query += ` WHERE creator=?`
		args = append(args, creator)
	}
	query += ` ORDER BY id DESC`
	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.FirewallRequest
	for rows.Next() {
		var r models.FirewallRequest
		if err := rows.Scan(&r.ID, &r.LegacyFormNumber, &r.SystemName, &r.Action, &r.PurposeType, &r.SourceZone, &r.SourceZone2, &r.SourceIP, &r.DestinationZone, &r.DestinationZone2, &r.DestinationIP, &r.ProtocolType, &r.StartDate, &r.EndDate, &r.RequestDate, &r.RuleDescription, &r.FirewallZone, &r.FirewallID, &r.Status, &r.Creator, &r.Remarks, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	if list == nil {
		list = []models.FirewallRequest{}
	}
	return list, nil
}

func (d *DB) GetFirewallRequest(id int) (*models.FirewallRequest, error) {
	return d.GetFirewallRequestByCreator(id, "")
}

func (d *DB) GetFirewallRequestByCreator(id int, creator string) (*models.FirewallRequest, error) {
	query := `SELECT id,legacy_form_number,system_name,action,purpose_type,source_zone,source_zone2,source_ip,destination_zone,destination_zone2,destination_ip,protocol_type,start_date,end_date,request_date,rule_description,firewall_zone,firewall_id,status,creator,remarks,created_at,updated_at FROM firewall_requests WHERE id=?`
	args := []interface{}{id}
	if strings.TrimSpace(creator) != "" {
		query += ` AND creator=?`
		args = append(args, creator)
	}
	row := d.conn.QueryRow(query, args...)
	var r models.FirewallRequest
	if err := row.Scan(&r.ID, &r.LegacyFormNumber, &r.SystemName, &r.Action, &r.PurposeType, &r.SourceZone, &r.SourceZone2, &r.SourceIP, &r.DestinationZone, &r.DestinationZone2, &r.DestinationIP, &r.ProtocolType, &r.StartDate, &r.EndDate, &r.RequestDate, &r.RuleDescription, &r.FirewallZone, &r.FirewallID, &r.Status, &r.Creator, &r.Remarks, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return nil, err
	}
	return &r, nil
}

func (d *DB) CreateFirewallRequest(r *models.FirewallRequest) (int64, error) {
	res, err := d.conn.Exec(`INSERT INTO firewall_requests (legacy_form_number,system_name,action,purpose_type,source_zone,source_zone2,source_ip,destination_zone,destination_zone2,destination_ip,protocol_type,start_date,end_date,request_date,rule_description,firewall_zone,firewall_id,status,creator,remarks) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		r.LegacyFormNumber, r.SystemName, r.Action, r.PurposeType, r.SourceZone, r.SourceZone2, r.SourceIP, r.DestinationZone, r.DestinationZone2, r.DestinationIP, r.ProtocolType, r.StartDate, r.EndDate, r.RequestDate, r.RuleDescription, r.FirewallZone, r.FirewallID, r.Status, r.Creator, r.Remarks)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) UpdateFirewallRequest(r *models.FirewallRequest) error {
	_, err := d.conn.Exec(`UPDATE firewall_requests SET legacy_form_number=?,system_name=?,action=?,purpose_type=?,source_zone=?,source_zone2=?,source_ip=?,destination_zone=?,destination_zone2=?,destination_ip=?,protocol_type=?,start_date=?,end_date=?,request_date=?,rule_description=?,firewall_zone=?,firewall_id=?,status=?,creator=?,remarks=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		r.LegacyFormNumber, r.SystemName, r.Action, r.PurposeType, r.SourceZone, r.SourceZone2, r.SourceIP, r.DestinationZone, r.DestinationZone2, r.DestinationIP, r.ProtocolType, r.StartDate, r.EndDate, r.RequestDate, r.RuleDescription, r.FirewallZone, r.FirewallID, r.Status, r.Creator, r.Remarks, r.ID)
	return err
}

func (d *DB) DeleteFirewallRequest(id int) error {
	return d.DeleteFirewallRequestByCreator(id, "")
}

func (d *DB) DeleteFirewallRequestByCreator(id int, creator string) error {
	query := `DELETE FROM firewall_requests WHERE id=?`
	args := []interface{}{id}
	if strings.TrimSpace(creator) != "" {
		query += ` AND creator=?`
		args = append(args, creator)
	}
	_, err := d.conn.Exec(query, args...)
	return err
}

func (d *DB) ListAssetInventoryRecords() ([]models.AssetInventoryRecord, error) {
	return d.ListAssetInventoryRecordsByCreator("", "all", "", "")
}

func (d *DB) ListAssetInventoryRecordsFiltered(keyword, status, assetType string) ([]models.AssetInventoryRecord, error) {
	return d.ListAssetInventoryRecordsByCreator(keyword, status, assetType, "")
}

func (d *DB) ListAssetInventoryRecordsByCreator(keyword, status, assetType, creator string) ([]models.AssetInventoryRecord, error) {
	base := `SELECT id,system_name,environment,asset_code,asset_type,asset_name,vendor_name,is_core_asset,has_national_security_concern,asset_description,quantity,os_config_baseline,browser_config_baseline,network_config_baseline,application_config_baseline,other_config_baseline,config_exception_code,manager_department,user_department,location,confidentiality,integrity,availability,asset_value,legal_compliance,protection_level,mtpd,rto,rpo,status,creator,remarks,created_at,updated_at FROM asset_inventory_records`
	clauses := []string{}
	args := []interface{}{}
	if strings.TrimSpace(status) != "" && status != "all" {
		clauses = append(clauses, "status = ?")
		args = append(args, status)
	}
	if strings.TrimSpace(assetType) != "" && assetType != "all" {
		clauses = append(clauses, "asset_type = ?")
		args = append(args, assetType)
	}
	if strings.TrimSpace(keyword) != "" {
		like := "%" + strings.TrimSpace(keyword) + "%"
		clauses = append(clauses, "(system_name LIKE ? OR environment LIKE ? OR asset_code LIKE ? OR asset_name LIKE ? OR manager_department LIKE ? OR user_department LIKE ? OR location LIKE ?)")
		args = append(args, like, like, like, like, like, like, like)
	}
	if strings.TrimSpace(creator) != "" {
		clauses = append(clauses, "creator = ?")
		args = append(args, creator)
	}
	query := base
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY id DESC"
	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.AssetInventoryRecord
	for rows.Next() {
		var r models.AssetInventoryRecord
		if err := rows.Scan(&r.ID, &r.SystemName, &r.Environment, &r.AssetCode, &r.AssetType, &r.AssetName, &r.VendorName, &r.IsCoreAsset, &r.HasNationalSecurityConcern, &r.AssetDescription, &r.Quantity, &r.OsConfigBaseline, &r.BrowserConfigBaseline, &r.NetworkConfigBaseline, &r.ApplicationConfigBaseline, &r.OtherConfigBaseline, &r.ConfigExceptionCode, &r.ManagerDepartment, &r.UserDepartment, &r.Location, &r.Confidentiality, &r.Integrity, &r.Availability, &r.AssetValue, &r.LegalCompliance, &r.ProtectionLevel, &r.Mtpd, &r.Rto, &r.Rpo, &r.Status, &r.Creator, &r.Remarks, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	if list == nil {
		list = []models.AssetInventoryRecord{}
	}
	return list, nil
}

func (d *DB) GetAssetInventoryRecord(id int) (*models.AssetInventoryRecord, error) {
	return d.GetAssetInventoryRecordByCreator(id, "")
}

func (d *DB) GetAssetInventoryRecordByCreator(id int, creator string) (*models.AssetInventoryRecord, error) {
	query := `SELECT id,system_name,environment,asset_code,asset_type,asset_name,vendor_name,is_core_asset,has_national_security_concern,asset_description,quantity,os_config_baseline,browser_config_baseline,network_config_baseline,application_config_baseline,other_config_baseline,config_exception_code,manager_department,user_department,location,confidentiality,integrity,availability,asset_value,legal_compliance,protection_level,mtpd,rto,rpo,status,creator,remarks,created_at,updated_at FROM asset_inventory_records WHERE id=?`
	args := []interface{}{id}
	if strings.TrimSpace(creator) != "" {
		query += ` AND creator=?`
		args = append(args, creator)
	}
	row := d.conn.QueryRow(query, args...)
	var r models.AssetInventoryRecord
	if err := row.Scan(&r.ID, &r.SystemName, &r.Environment, &r.AssetCode, &r.AssetType, &r.AssetName, &r.VendorName, &r.IsCoreAsset, &r.HasNationalSecurityConcern, &r.AssetDescription, &r.Quantity, &r.OsConfigBaseline, &r.BrowserConfigBaseline, &r.NetworkConfigBaseline, &r.ApplicationConfigBaseline, &r.OtherConfigBaseline, &r.ConfigExceptionCode, &r.ManagerDepartment, &r.UserDepartment, &r.Location, &r.Confidentiality, &r.Integrity, &r.Availability, &r.AssetValue, &r.LegalCompliance, &r.ProtectionLevel, &r.Mtpd, &r.Rto, &r.Rpo, &r.Status, &r.Creator, &r.Remarks, &r.CreatedAt, &r.UpdatedAt); err != nil {
		return nil, err
	}
	return &r, nil
}

func (d *DB) CreateAssetInventoryRecord(r *models.AssetInventoryRecord) (int64, error) {
	res, err := d.conn.Exec(`INSERT INTO asset_inventory_records (system_name,environment,asset_code,asset_type,asset_name,vendor_name,is_core_asset,has_national_security_concern,asset_description,quantity,os_config_baseline,browser_config_baseline,network_config_baseline,application_config_baseline,other_config_baseline,config_exception_code,manager_department,user_department,location,confidentiality,integrity,availability,asset_value,legal_compliance,protection_level,mtpd,rto,rpo,status,creator,remarks) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		r.SystemName, r.Environment, r.AssetCode, r.AssetType, r.AssetName, r.VendorName, r.IsCoreAsset, r.HasNationalSecurityConcern, r.AssetDescription, r.Quantity, r.OsConfigBaseline, r.BrowserConfigBaseline, r.NetworkConfigBaseline, r.ApplicationConfigBaseline, r.OtherConfigBaseline, r.ConfigExceptionCode, r.ManagerDepartment, r.UserDepartment, r.Location, r.Confidentiality, r.Integrity, r.Availability, r.AssetValue, r.LegalCompliance, r.ProtectionLevel, r.Mtpd, r.Rto, r.Rpo, r.Status, r.Creator, r.Remarks)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) UpdateAssetInventoryRecord(r *models.AssetInventoryRecord) error {
	_, err := d.conn.Exec(`UPDATE asset_inventory_records SET system_name=?,environment=?,asset_code=?,asset_type=?,asset_name=?,vendor_name=?,is_core_asset=?,has_national_security_concern=?,asset_description=?,quantity=?,os_config_baseline=?,browser_config_baseline=?,network_config_baseline=?,application_config_baseline=?,other_config_baseline=?,config_exception_code=?,manager_department=?,user_department=?,location=?,confidentiality=?,integrity=?,availability=?,asset_value=?,legal_compliance=?,protection_level=?,mtpd=?,rto=?,rpo=?,status=?,creator=?,remarks=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		r.SystemName, r.Environment, r.AssetCode, r.AssetType, r.AssetName, r.VendorName, r.IsCoreAsset, r.HasNationalSecurityConcern, r.AssetDescription, r.Quantity, r.OsConfigBaseline, r.BrowserConfigBaseline, r.NetworkConfigBaseline, r.ApplicationConfigBaseline, r.OtherConfigBaseline, r.ConfigExceptionCode, r.ManagerDepartment, r.UserDepartment, r.Location, r.Confidentiality, r.Integrity, r.Availability, r.AssetValue, r.LegalCompliance, r.ProtectionLevel, r.Mtpd, r.Rto, r.Rpo, r.Status, r.Creator, r.Remarks, r.ID)
	return err
}

func (d *DB) DeleteAssetInventoryRecord(id int) error {
	return d.DeleteAssetInventoryRecordByCreator(id, "")
}

func (d *DB) DeleteAssetInventoryRecordByCreator(id int, creator string) error {
	query := `DELETE FROM asset_inventory_records WHERE id=?`
	args := []interface{}{id}
	if strings.TrimSpace(creator) != "" {
		query += ` AND creator=?`
		args = append(args, creator)
	}
	_, err := d.conn.Exec(query, args...)
	return err
}
