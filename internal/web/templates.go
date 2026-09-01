package web

import (
	"html/template"
	"net/http"
	"path/filepath"
)

type NavItem struct {
	Label    string
	Icon     string
	Href     string
	IsActive bool
}

type PageData struct {
	Title    string
	NavItems []NavItem
}

type Renderer struct {
	templates *template.Template
}

func NewRenderer(templateRoot string) (*Renderer, error) {
	pattern := filepath.Join(templateRoot, "*.gohtml")
	tpls, err := template.ParseGlob(pattern)
	if err != nil {
		return nil, err
	}
	return &Renderer{templates: tpls}, nil
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data PageData) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return r.templates.ExecuteTemplate(w, name, data)
}

func BuildNav(activeHref string) []NavItem {
	items := []NavItem{
		{Label: "儀表板", Icon: "📊", Href: "/"},
		{Label: "04-042防火牆申請", Icon: "🔥", Href: "/firewall-requests"},
		{Label: "04-008資訊資產清冊", Icon: "📦", Href: "/asset-inventory"},
		{Label: "04-062特殊權限帳號管理", Icon: "👤", Href: "/accounts"},
		{Label: "04-069防護基準執行說明", Icon: "🛡️", Href: "/protection-baselines"},
		{Label: "04-078系統平台申請", Icon: "🖥️", Href: "/platform-requests"},
		{Label: "表單管理", Icon: "🗂️", Href: "/"},
		{Label: "通知記錄", Icon: "📧", Href: "/"},
		{Label: "操作日誌", Icon: "🧾", Href: "/"},
	}
	for i := range items {
		items[i].IsActive = items[i].Href == activeHref
	}
	return items
}
