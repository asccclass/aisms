package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// key 型別避免 context key 衝突
type contextKey string

const userEmailKey contextKey = "user_email"

var oauth2Config *oauth2.Config

type googleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
	HD            string `json:"hd"`
}

type sessionUserProfile struct {
	Department       string
	DepartmentSource string
	DepartmentNote   string
	Title            string
	OrganizationName string
	OrgUnitPath      string
}

func InitOAuth2() {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URI")
	if clientID == "" || clientSecret == "" {
		fmt.Println("Warning: GOOGLE_CLIENT_ID or SECRET not set")
		return
	}
	oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func isDevMockLoginEnabled() bool {
	return os.Getenv("ENABLE_DEV_MOCK_LOGIN") == "true"
}

func getAllowedGoogleDomains() []string {
	raw := strings.TrimSpace(os.Getenv("GOOGLE_ALLOWED_DOMAINS"))
	if raw == "" {
		return nil
	}
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\n' || r == '\t'
	})
	list := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.ToLower(strings.TrimSpace(field))
		if field != "" {
			list = append(list, field)
		}
	}
	return list
}

func inferDepartment(ui googleUserInfo) string {
	if mapped := lookupDepartmentByDomain(ui.HD); mapped != "" {
		return mapped
	}
	emailDomain := emailDomain(ui.Email)
	if mapped := lookupDepartmentByDomain(emailDomain); mapped != "" {
		return mapped
	}
	if ui.HD != "" {
		return ui.HD
	}
	if emailDomain != "" {
		return emailDomain
	}
	return ""
}

func lookupDepartmentByDomain(domain string) string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return ""
	}
	raw := strings.TrimSpace(os.Getenv("GOOGLE_DOMAIN_DEPARTMENTS"))
	if raw == "" {
		return ""
	}
	pairs := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ';' || r == '\n'
	})
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		if key == domain && value != "" {
			return value
		}
	}
	return ""
}

func emailDomain(email string) string {
	parts := strings.Split(strings.TrimSpace(email), "@")
	if len(parts) != 2 {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(parts[1]))
}

func isAllowedHostedDomain(ui googleUserInfo) bool {
	allowed := getAllowedGoogleDomains()
	if len(allowed) == 0 {
		return true
	}
	candidates := []string{strings.ToLower(strings.TrimSpace(ui.HD)), emailDomain(ui.Email)}
	for _, candidate := range candidates {
		for _, domain := range allowed {
			if candidate != "" && candidate == domain {
				return true
			}
		}
	}
	return false
}

func setSessionCookie(w http.ResponseWriter, name, value string, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: httpOnly,
	})
}

func expireCookie(w http.ResponseWriter, name string, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: httpOnly,
	})
}

func writeGoogleSessionCookies(w http.ResponseWriter, ui googleUserInfo, profile sessionUserProfile) {
	setSessionCookie(w, "admin_email", ui.Email, true)
	setSessionCookie(w, "admin_name", ui.Name, false)
	setSessionCookie(w, "admin_picture", ui.Picture, false)
	setSessionCookie(w, "admin_hd", ui.HD, false)
	setSessionCookie(w, "admin_google_id", ui.ID, true)
	setSessionCookie(w, "admin_locale", ui.Locale, false)
	setSessionCookie(w, "admin_given_name", ui.GivenName, false)
	setSessionCookie(w, "admin_family_name", ui.FamilyName, false)
	setSessionCookie(w, "admin_verified_email", fmt.Sprintf("%t", ui.VerifiedEmail), false)
	setSessionCookie(w, "admin_department", profile.Department, false)
	setSessionCookie(w, "admin_department_source", profile.DepartmentSource, false)
	setSessionCookie(w, "admin_department_note", profile.DepartmentNote, false)
	setSessionCookie(w, "admin_title", profile.Title, false)
	setSessionCookie(w, "admin_organization_name", profile.OrganizationName, false)
	setSessionCookie(w, "admin_org_unit_path", profile.OrgUnitPath, false)
}

func buildSessionUserProfile(ui googleUserInfo, resolved workspaceProfile) sessionUserProfile {
	department := firstNonEmpty(resolved.Department, inferDepartment(ui))
	source := firstNonEmpty(resolved.Source, "hd")
	note := "Google basic userinfo 不會直接提供單位；系統會依序嘗試 Directory API、People API，最後才 fallback 到 hosted domain / email domain。"
	if source == "directory" || source == "directory_custom_schema" {
		note = "單位資訊來自 Google Workspace Directory API。"
	} else if source == "people" {
		note = "單位資訊來自 Google People API 的 organizations 欄位。"
	} else if source == "hd" {
		note = "單位資訊無法直接從 Google Workspace 正式欄位取得，目前使用 hosted domain / email domain 或對照表推估。"
	}
	return sessionUserProfile{
		Department:       department,
		DepartmentSource: source,
		DepartmentNote:   note,
		Title:            strings.TrimSpace(resolved.Title),
		OrganizationName: strings.TrimSpace(resolved.OrganizationName),
		OrgUnitPath:      strings.TrimSpace(resolved.OrgUnitPath),
	}
}

// 登入介面
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if oauth2Config == nil {
		http.Error(w, "OAuth not configured", 503)
		return
	}
	mockEmail := r.URL.Query().Get("mock_email")
	if mockEmail != "" && isDevMockLoginEnabled() {
		http.SetCookie(w, &http.Cookie{
			Name:     "admin_email",
			Value:    mockEmail,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})
		http.SetCookie(w, &http.Cookie{
			Name:     "admin_name",
			Value:    "測試人員",
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: false, // allow JS to read
		})
		mockUI := googleUserInfo{
			Email:         mockEmail,
			Name:          "測試人員",
			GivenName:     "測試",
			FamilyName:    "人員",
			Locale:        "zh-TW",
			HD:            emailDomain(mockEmail),
			VerifiedEmail: true,
		}
		writeGoogleSessionCookies(w, mockUI, buildSessionUserProfile(mockUI, workspaceProfile{
			Department: inferDepartment(mockUI),
			Source:     "hd",
		}))
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	authOpts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOnline,
		oauth2.SetAuthURLParam("prompt", "select_account"),
	}
	if allowed := getAllowedGoogleDomains(); len(allowed) > 0 {
		authOpts = append(authOpts, oauth2.SetAuthURLParam("hd", allowed[0]))
	}
	url := oauth2Config.AuthCodeURL("state-token", authOpts...)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Google callback
func (h *Handler) OAuth2Callback(w http.ResponseWriter, r *http.Request) {
	if oauth2Config == nil {
		http.Error(w, "OAuth not configured", 503)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		http.Redirect(w, r, "/?error=missing_code", http.StatusTemporaryRedirect)
		return
	}

	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		http.Redirect(w, r, "/?error=exchange_failed", http.StatusTemporaryRedirect)
		return
	}

	client := oauth2Config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil || resp.StatusCode != 200 {
		http.Redirect(w, r, "/?error=userinfo_failed", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	var ui googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&ui); err != nil {
		http.Redirect(w, r, "/?error=userinfo_decode_failed", http.StatusTemporaryRedirect)
		return
	}

	if ui.Email == "" {
		http.Redirect(w, r, "/?error=no_email", http.StatusTemporaryRedirect)
		return
	}
	if !ui.VerifiedEmail {
		http.Redirect(w, r, "/?error=email_not_verified", http.StatusTemporaryRedirect)
		return
	}
	if !isAllowedHostedDomain(ui) {
		http.Redirect(w, r, "/?error=domain_not_allowed", http.StatusTemporaryRedirect)
		return
	}

	resolvedProfile := resolveWorkspaceProfile(context.Background(), ui)
	writeGoogleSessionCookies(w, ui, buildSessionUserProfile(ui, resolvedProfile))

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// 登出
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	expireCookie(w, "admin_email", true)
	expireCookie(w, "admin_name", false)
	expireCookie(w, "admin_picture", false)
	expireCookie(w, "admin_hd", false)
	expireCookie(w, "admin_google_id", true)
	expireCookie(w, "admin_locale", false)
	expireCookie(w, "admin_given_name", false)
	expireCookie(w, "admin_family_name", false)
	expireCookie(w, "admin_verified_email", false)
	expireCookie(w, "admin_department", false)
	expireCookie(w, "admin_department_source", false)
	expireCookie(w, "admin_department_note", false)
	expireCookie(w, "admin_title", false)
	expireCookie(w, "admin_organization_name", false)
	expireCookie(w, "admin_org_unit_path", false)
	writeJSON(w, 200, map[string]string{"message": "登出成功"})
}

// 取得當前使用者
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	emailCookie, err := r.Cookie("admin_email")
	if err != nil || emailCookie.Value == "" {
		writeJSON(w, 401, map[string]string{"error": "Unauthorized"})
		return
	}

	cookieValue := func(name string) string {
		c, err := r.Cookie(name)
		if err != nil || c == nil {
			return ""
		}
		return c.Value
	}

	writeJSON(w, 200, map[string]interface{}{
		"email":             emailCookie.Value,
		"name":              cookieValue("admin_name"),
		"given_name":        cookieValue("admin_given_name"),
		"family_name":       cookieValue("admin_family_name"),
		"picture":           cookieValue("admin_picture"),
		"hosted_domain":     cookieValue("admin_hd"),
		"google_id":         cookieValue("admin_google_id"),
		"locale":            cookieValue("admin_locale"),
		"verified_email":    cookieValue("admin_verified_email") == "true",
		"email_domain":      emailDomain(emailCookie.Value),
		"department":        cookieValue("admin_department"),
		"department_source": cookieValue("admin_department_source"),
		"department_note":   cookieValue("admin_department_note"),
		"title":             cookieValue("admin_title"),
		"organization_name": cookieValue("admin_organization_name"),
		"org_unit_path":     cookieValue("admin_org_unit_path"),
	})
}

// AuthMiddleware 驗證登入狀態
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("admin_email")
		if err != nil || c.Value == "" {
			writeJSON(w, 401, map[string]string{"error": "Unauthorized"})
			return
		}
		// 將 email 放進 context 供後續使用
		ctx := context.WithValue(r.Context(), userEmailKey, c.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserEmail 從 context 取出
func GetUserEmail(r *http.Request) string {
	val := r.Context().Value(userEmailKey)
	if val != nil {
		return val.(string)
	}
	return ""
}
