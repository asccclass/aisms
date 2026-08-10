package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"isms-privilege/internal/db"
	"isms-privilege/internal/mailer"
	"isms-privilege/internal/models"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	DB     *db.DB
	Mailer *mailer.Mailer
}

var dashboardProviders = []models.DashboardFormProvider{
	{Key: "privileged_accounts", Label: "特殊權限帳號資料", Description: "使用 privileged_accounts 資料表作為首頁表單資料來源"},
	{Key: "placeholder", Label: "示範骨架 / 尚未接資料", Description: "保留表單卡片與說明，首頁顯示空狀態"},
	{Key: "custom_table_template", Label: "自訂資料表範本", Description: "作為未來接新資料表的 provider 樣板，預設先回傳空資料"},
}

func New(d *db.DB, m *mailer.Mailer) *Handler {
	return &Handler{DB: d, Mailer: m}
}

// ---- helpers ----

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func idFromPath(r *http.Request, prefix string) (int, error) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, prefix), "/")
	if len(parts) == 0 || parts[0] == "" {
		return 0, fmt.Errorf("missing id")
	}
	return strconv.Atoi(parts[0])
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ---- Accounts CRUD ----

// GET /api/accounts?status=active&q=keyword
func (h *Handler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	q := r.URL.Query().Get("q")
	var (
		accounts []models.PrivilegedAccount
		err      error
	)
	if q != "" {
		accounts, err = h.DB.SearchAccounts(q)
	} else {
		accounts, err = h.DB.ListAccounts(status)
	}
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if accounts == nil {
		accounts = []models.PrivilegedAccount{}
	}
	writeJSON(w, 200, accounts)
}

// GET /api/accounts/{id}
func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/accounts/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	a, err := h.DB.GetAccount(id)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, 200, a)
}

// POST /api/accounts
func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var a models.PrivilegedAccount
	if err := readJSON(r, &a); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	a.Creator = GetUserEmail(r)
	if a.Status == "" {
		a.Status = models.StatusActive
	}
	id, err := h.DB.CreateAccount(&a)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	a.ID = int(id)
	writeJSON(w, 201, a)
}

// PUT /api/accounts/{id}
func (h *Handler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/accounts/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	var a models.PrivilegedAccount
	if err := readJSON(r, &a); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	a.ID = id
	existing, _ := h.DB.GetAccount(id)
	if existing != nil {
		a.Creator = existing.Creator
	}
	if err := h.DB.UpdateAccount(&a); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, a)
}

// DELETE /api/accounts/{id}
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/accounts/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.DB.DeleteAccount(id); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"message": "deleted"})
}

// ---- Email Notification ----

// POST /api/accounts/{id}/notify  → send single notification
func (h *Handler) NotifyAccount(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/accounts/"), "/notify")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	a, err := h.DB.GetAccount(id)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	if a.Email == "" {
		writeJSON(w, 400, map[string]string{"error": "no email address"})
		return
	}
	token, err := generateToken()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "token generation failed"})
		return
	}
	expiry := time.Now().AddDate(0, 0, 7)
	if err := h.DB.SetConfirmToken(id, token, expiry); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	logStatus := "sent"
	logMsg := "通知發送成功"
	if err := h.Mailer.SendConfirmEmail(a.Email, a.OwnerName, a.SystemName, a.IPAddress, a.AccountName, a.AccountType, token); err != nil {
		logStatus = "failed"
		logMsg = err.Error()
	}
	h.DB.AddNotificationLog(&models.NotificationLog{
		AccountID:   id,
		AccountName: a.AccountName,
		Email:       a.Email,
		Status:      logStatus,
		Message:     logMsg,
	})
	if logStatus == "failed" {
		writeJSON(w, 500, map[string]string{"error": logMsg, "debug": "token_set_ok"})
		return
	}
	writeJSON(w, 200, map[string]string{"message": "notification sent", "email": a.Email})
}

// POST /api/notify-all  → bulk notify all active accounts with email
func (h *Handler) NotifyAll(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.DB.GetActiveAccountsWithEmail()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	sent, failed := 0, 0
	for _, a := range accounts {
		token, err := generateToken()
		if err != nil {
			failed++
			continue
		}
		expiry := time.Now().AddDate(0, 0, 7)
		h.DB.SetConfirmToken(a.ID, token, expiry)
		logStatus := "sent"
		logMsg := "通知發送成功"
		if err := h.Mailer.SendConfirmEmail(a.Email, a.OwnerName, a.SystemName, a.IPAddress, a.AccountName, a.AccountType, token); err != nil {
			logStatus = "failed"
			logMsg = err.Error()
			failed++
		} else {
			sent++
		}
		h.DB.AddNotificationLog(&models.NotificationLog{
			AccountID:   a.ID,
			AccountName: a.AccountName,
			Email:       a.Email,
			Status:      logStatus,
			Message:     logMsg,
		})
	}
	writeJSON(w, 200, map[string]int{"sent": sent, "failed": failed, "total": len(accounts)})
}

// ---- Confirm Page (user-facing) ----

// GET /confirm?token=xxx&action=continue|stop
func (h *Handler) ConfirmAccount(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	action := r.URL.Query().Get("action")
	if token == "" || (action != "continue" && action != "stop") {
		http.Error(w, "無效的確認連結", 400)
		return
	}
	a, err := h.DB.GetAccountByToken(token)
	if err != nil {
		http.Error(w, "連結已失效或不存在，請聯繫管理員。", 404)
		return
	}
	if a.TokenExpiry != nil && a.TokenExpiry.Before(time.Now()) {
		http.Error(w, "確認連結已過期，請聯繫管理員重新發送。", 400)
		return
	}
	if err := h.DB.ConfirmAccount(a.ID, action); err != nil {
		http.Error(w, "系統錯誤，請稍後再試。", 500)
		return
	}
	actionText := "繼續使用"
	actionColor := "#28a745"
	if action == "stop" {
		actionText = "停用帳號"
		actionColor = "#dc3545"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="zh-TW"><head><meta charset="UTF-8"><title>確認完成</title>
<style>body{font-family:Arial,sans-serif;display:flex;justify-content:center;align-items:center;min-height:100vh;margin:0;background:#f4f6f8;}
.box{background:#fff;border-radius:12px;padding:40px;text-align:center;box-shadow:0 4px 20px rgba(0,0,0,.1);max-width:480px;width:90%%;}
.icon{font-size:60px;}.title{font-size:24px;font-weight:bold;color:%s;margin:16px 0 8px;}
.desc{color:#666;line-height:1.6;}.account{background:#f8f9fa;border-radius:8px;padding:12px;margin:16px 0;font-size:14px;}
a{display:inline-block;margin-top:20px;background:#1a3a5c;color:#fff;padding:10px 24px;border-radius:8px;text-decoration:none;}</style></head>
<body><div class="box">
<div class="icon">%s</div>
<div class="title">確認完成 — %s</div>
<p class="desc">您已成功完成特殊權限帳號盤點確認。</p>
<div class="account">系統：%s &nbsp;|&nbsp; 帳號：%s &nbsp;|&nbsp; 使用者：%s</div>
<p class="desc" style="font-size:13px;color:#888;">確認結果已記錄至 ISMS 系統，感謝您的配合。</p>
<a href="/">返回首頁</a></div></body></html>`,
		actionColor,
		map[string]string{"continue": "✅", "stop": "🔒"}[action],
		actionText,
		a.SystemName, a.AccountName, a.OwnerName)
}

// ---- Stats ----

// GET /api/stats
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.DB.Stats()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, stats)
}

// ---- Notification Logs ----

// GET /api/notification-logs
func (h *Handler) ListLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.DB.ListNotificationLogs()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	if logs == nil {
		logs = []models.NotificationLog{}
	}
	writeJSON(w, 200, logs)
}

// ---- Dashboard Forms ----

func (h *Handler) ListDashboardForms(w http.ResponseWriter, r *http.Request) {
	forms, err := h.DB.ListDashboardForms()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, forms)
}

func (h *Handler) GetDashboardForm(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/dashboard-forms/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	form, err := h.DB.GetDashboardForm(id)
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, 200, form)
}

func (h *Handler) CreateDashboardForm(w http.ResponseWriter, r *http.Request) {
	var form models.DashboardForm
	if err := readJSON(r, &form); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	if strings.TrimSpace(form.Key) == "" || strings.TrimSpace(form.Code) == "" || strings.TrimSpace(form.Name) == "" {
		writeJSON(w, 400, map[string]string{"error": "key, code, name are required"})
		return
	}
	id, err := h.DB.CreateDashboardForm(&form)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	form.ID = int(id)
	writeJSON(w, 201, form)
}

func (h *Handler) UpdateDashboardForm(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/dashboard-forms/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	var form models.DashboardForm
	if err := readJSON(r, &form); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	form.ID = id
	if strings.TrimSpace(form.Key) == "" || strings.TrimSpace(form.Code) == "" || strings.TrimSpace(form.Name) == "" {
		writeJSON(w, 400, map[string]string{"error": "key, code, name are required"})
		return
	}
	if err := h.DB.UpdateDashboardForm(&form); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, form)
}

func (h *Handler) DeleteDashboardForm(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/dashboard-forms/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	if err := h.DB.DeleteDashboardForm(id); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"message": "deleted"})
}

func (h *Handler) ReorderDashboardForms(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		IDs []int `json:"ids"`
	}
	if err := readJSON(r, &payload); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	if len(payload.IDs) == 0 {
		writeJSON(w, 400, map[string]string{"error": "ids are required"})
		return
	}
	if err := h.DB.UpdateDashboardFormOrder(payload.IDs); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"message": "reordered"})
}

func (h *Handler) ListDashboardProviders(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, dashboardProviders)
}

func (h *Handler) ListDashboardRecords(w http.ResponseWriter, r *http.Request) {
	formKey := strings.TrimSpace(r.URL.Query().Get("form_key"))
	if formKey == "" {
		writeJSON(w, 400, map[string]string{"error": "form_key is required"})
		return
	}
	forms, err := h.DB.ListDashboardForms()
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	var form *models.DashboardForm
	for i := range forms {
		if forms[i].Key == formKey {
			form = &forms[i]
			break
		}
	}
	if form == nil {
		writeJSON(w, 404, map[string]string{"error": "dashboard form not found"})
		return
	}
	records, err := h.loadDashboardRecords(*form)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, records)
}

// loadDashboardRecords 依 form.ProviderKey 分派對應 provider。
//
// Provider 開發指南：
// 1. 在 internal/models 新增 provider 專屬資料 struct。
// 2. 在 internal/db 新增 ListXxxRecords() 查詢函式。
// 3. 在這裡的 switch 補一個 case，呼叫 loadXxxDashboardRecords()。
// 4. 在 loadXxxDashboardRecords() 內把 provider 專屬資料映射成 []models.DashboardRecord。
// 5. 在 dashboardProviders 清單加入新 provider，前端表單管理頁就能選到。
//
// 範例：
//   case "asset_inventory":
//     return h.loadAssetInventoryDashboardRecords()
func (h *Handler) loadDashboardRecords(form models.DashboardForm) ([]models.DashboardRecord, error) {
	switch form.ProviderKey {
	case "privileged_accounts":
		return h.loadPrivilegedAccountDashboardRecords()
	case "placeholder":
		return h.loadPlaceholderDashboardRecords()
	case "custom_table_template":
		return h.loadCustomTableTemplateDashboardRecords()
	// case "asset_inventory":
	// 	return h.loadAssetInventoryDashboardRecords()
	default:
		return []models.DashboardRecord{}, nil
	}
}

func (h *Handler) loadPrivilegedAccountDashboardRecords() ([]models.DashboardRecord, error) {
	accounts, err := h.DB.ListAccounts("all")
	if err != nil {
		return nil, err
	}
	records := make([]models.DashboardRecord, 0, len(accounts))
	for _, a := range accounts {
		records = append(records, models.DashboardRecord{
			ID:            a.ID,
			PrimaryName:   a.AccountName,
			SecondaryName: fmt.Sprintf("%s / %s", a.SystemName, a.Environment),
			OwnerName:     a.OwnerName,
			Status:        string(a.Status),
			InventoryDate: a.InventoryDate,
			UpdatedAt:     a.UpdatedAt,
			Email:         a.Email,
		})
	}
	return records, nil
}

func (h *Handler) loadPlaceholderDashboardRecords() ([]models.DashboardRecord, error) {
	return []models.DashboardRecord{}, nil
}

func (h *Handler) loadCustomTableTemplateDashboardRecords() ([]models.DashboardRecord, error) {
	rows, err := h.DB.ListCustomTableTemplateRecords()
	if err != nil {
		return nil, err
	}
	records := make([]models.DashboardRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, models.DashboardRecord{
			ID:            row.ID,
			PrimaryName:   row.Title,
			SecondaryName: row.Category,
			OwnerName:     row.OwnerName,
			Status:        row.Status,
			InventoryDate: row.InventoryDate,
			UpdatedAt:     row.UpdatedAt,
			Email:         row.Email,
		})
	}
	return records, nil
}

// loadAssetInventoryDashboardRecords 是新增 provider 時可直接照抄的範例骨架。
//
// 啟用方式：
// 1. 先在 models 新增 AssetInventoryRecord struct
// 2. 再在 db 新增 ListAssetInventoryRecords()
// 3. 最後打開 loadDashboardRecords() 裡的 case "asset_inventory"
//
// 這裡先保留註解骨架，避免現在就引入尚未存在的資料表依賴。
//
// func (h *Handler) loadAssetInventoryDashboardRecords() ([]models.DashboardRecord, error) {
// 	rows, err := h.DB.ListAssetInventoryRecords()
// 	if err != nil {
// 		return nil, err
// 	}
// 	records := make([]models.DashboardRecord, 0, len(rows))
// 	for _, row := range rows {
// 		records = append(records, models.DashboardRecord{
// 			ID:            row.ID,
// 			PrimaryName:   row.AssetName,
// 			SecondaryName: fmt.Sprintf("%s / %s", row.AssetType, row.Department),
// 			OwnerName:     row.OwnerName,
// 			Status:        row.Status,
// 			InventoryDate: row.InventoryDate,
// 			UpdatedAt:     row.UpdatedAt,
// 			Email:         row.Email,
// 		})
// 	}
// 	return records, nil
// }

// RegisterRoutes 掛載所有路由到 mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// SPA & confirm page
	mux.HandleFunc("/confirm", h.ConfirmAccount)

	// Auth
	mux.HandleFunc("/api/login", h.Login)
	mux.HandleFunc("/api/logout", h.Logout)
	mux.HandleFunc("/oauth2callback", h.OAuth2Callback)
	mux.HandleFunc("/api/me", h.Me)

	// API
	mux.HandleFunc("/api/stats", AuthMiddleware(h.Stats))
	mux.HandleFunc("/api/notification-logs", AuthMiddleware(h.ListLogs))
	mux.HandleFunc("/api/notify-all", AuthMiddleware(h.NotifyAll))
	mux.HandleFunc("/api/dashboard-providers", AuthMiddleware(h.ListDashboardProviders))
	mux.HandleFunc("/api/dashboard-records", AuthMiddleware(h.ListDashboardRecords))
	mux.HandleFunc("/api/dashboard-forms/reorder", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.ReorderDashboardForms(w, r)
			return
		}
		http.Error(w, "method not allowed", 405)
	}))
	mux.HandleFunc("/api/dashboard-forms", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListDashboardForms(w, r)
		case http.MethodPost:
			h.CreateDashboardForm(w, r)
		default:
			http.Error(w, "method not allowed", 405)
		}
	}))
	mux.HandleFunc("/api/dashboard-forms/", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetDashboardForm(w, r)
		case http.MethodPut:
			h.UpdateDashboardForm(w, r)
		case http.MethodDelete:
			h.DeleteDashboardForm(w, r)
		default:
			http.Error(w, "method not allowed", 405)
		}
	}))

	mux.HandleFunc("/api/accounts", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.ListAccounts(w, r)
		case http.MethodPost:
			h.CreateAccount(w, r)
		default:
			http.Error(w, "method not allowed", 405)
		}
	}))

	mux.HandleFunc("/api/accounts/", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/notify") {
			if r.Method == http.MethodPost {
				h.NotifyAccount(w, r)
			} else {
				http.Error(w, "method not allowed", 405)
			}
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.GetAccount(w, r)
		case http.MethodPut:
			h.UpdateAccount(w, r)
		case http.MethodDelete:
			h.DeleteAccount(w, r)
		default:
			http.Error(w, "method not allowed", 405)
		}
	}))
}
