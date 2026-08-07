package db

import (
	"database/sql"
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
	`
	_, err := d.conn.Exec(schema)
	_, _ = d.conn.Exec(`ALTER TABLE privileged_accounts ADD COLUMN environment TEXT NOT NULL DEFAULT '正式區'`)
	_, _ = d.conn.Exec(`ALTER TABLE privileged_accounts ADD COLUMN department_code TEXT NOT NULL DEFAULT ''`)
	_, _ = d.conn.Exec(`ALTER TABLE privileged_accounts ADD COLUMN creator TEXT NOT NULL DEFAULT ''`)
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
