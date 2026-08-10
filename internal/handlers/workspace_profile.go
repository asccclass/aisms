package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
)

type workspaceProfile struct {
	Department       string
	Title            string
	OrganizationName string
	OrgUnitPath      string
	Source           string
}

type directoryUser struct {
	OrgUnitPath   string `json:"orgUnitPath"`
	Organizations []struct {
		Name       string `json:"name"`
		Department string `json:"department"`
		Title      string `json:"title"`
		Primary    bool   `json:"primary"`
		Type       string `json:"type"`
	} `json:"organizations"`
	CustomSchemas map[string]map[string]interface{} `json:"customSchemas"`
}

type peoplePerson struct {
	Organizations []struct {
		Name       string `json:"name"`
		Department string `json:"department"`
		Title      string `json:"title"`
		Current    bool   `json:"current"`
		Type       string `json:"type"`
		Domain     string `json:"domain"`
	} `json:"organizations"`
}

func resolveWorkspaceProfile(ctx context.Context, ui googleUserInfo) workspaceProfile {
	if profile, err := resolveDirectoryProfile(ctx, ui.Email); err == nil && profile.hasUsefulData() {
		return profile
	}
	if profile, err := resolvePeopleProfile(ctx, ui.Email); err == nil && profile.hasUsefulData() {
		profile.Source = "people"
		return profile
	}
	department := inferDepartment(ui)
	if department == "" {
		department = firstNonEmpty(ui.HD, emailDomain(ui.Email))
	}
	return workspaceProfile{
		Department: department,
		Source:     "hd",
	}
}

func (p workspaceProfile) hasUsefulData() bool {
	return strings.TrimSpace(p.Department) != "" ||
		strings.TrimSpace(p.Title) != "" ||
		strings.TrimSpace(p.OrganizationName) != "" ||
		strings.TrimSpace(p.OrgUnitPath) != ""
}

func resolveDirectoryProfile(ctx context.Context, userEmail string) (workspaceProfile, error) {
	client, err := newWorkspaceClient(ctx, os.Getenv("GOOGLE_WORKSPACE_ADMIN_EMAIL"), []string{
		"https://www.googleapis.com/auth/admin.directory.user.readonly",
	})
	if err != nil {
		return workspaceProfile{}, err
	}

	endpoint := "https://admin.googleapis.com/admin/directory/v1/users/" + url.PathEscape(userEmail) + "?projection=full&viewType=admin_view"
	var du directoryUser
	if err := doGoogleJSONRequest(ctx, client, endpoint, &du); err != nil {
		return workspaceProfile{}, err
	}

	profile := workspaceProfile{
		OrgUnitPath: strings.TrimSpace(du.OrgUnitPath),
		Source:      "directory",
	}

	if dept, title, orgName := findDirectoryCustomSchemaFields(du.CustomSchemas); dept != "" || title != "" || orgName != "" {
		profile.Department = dept
		profile.Title = title
		profile.OrganizationName = orgName
		if profile.Department != "" {
			profile.Source = "directory_custom_schema"
		}
	}

	if profile.Department == "" || profile.Title == "" || profile.OrganizationName == "" {
		dept, title, orgName := selectDirectoryOrganization(du.Organizations)
		profile.Department = firstNonEmpty(profile.Department, dept)
		profile.Title = firstNonEmpty(profile.Title, title)
		profile.OrganizationName = firstNonEmpty(profile.OrganizationName, orgName)
	}

	if profile.Department == "" {
		profile.Department = strings.Trim(strings.TrimSpace(profile.OrgUnitPath), "/")
	}

	return profile, nil
}

func resolvePeopleProfile(ctx context.Context, userEmail string) (workspaceProfile, error) {
	client, err := newWorkspaceClient(ctx, userEmail, []string{
		"https://www.googleapis.com/auth/user.organization.read",
	})
	if err != nil {
		return workspaceProfile{}, err
	}

	endpoint := "https://people.googleapis.com/v1/people/me?personFields=organizations"
	var person peoplePerson
	if err := doGoogleJSONRequest(ctx, client, endpoint, &person); err != nil {
		return workspaceProfile{}, err
	}

	dept, title, orgName := selectPeopleOrganization(person.Organizations)
	return workspaceProfile{
		Department:       dept,
		Title:            title,
		OrganizationName: orgName,
		Source:           "people",
	}, nil
}

func newWorkspaceClient(ctx context.Context, subject string, scopes []string) (*http.Client, error) {
	saJSON, err := readWorkspaceServiceAccountJSON()
	if err != nil {
		return nil, err
	}
	cfg, err := google.JWTConfigFromJSON(saJSON, scopes...)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(subject) == "" {
		return nil, fmt.Errorf("workspace subject email is required")
	}
	cfg.Subject = strings.TrimSpace(subject)
	return cfg.Client(ctx), nil
}

func readWorkspaceServiceAccountJSON() ([]byte, error) {
	if raw := strings.TrimSpace(os.Getenv("GOOGLE_WORKSPACE_SERVICE_ACCOUNT_JSON")); raw != "" {
		return []byte(raw), nil
	}
	if path := strings.TrimSpace(os.Getenv("GOOGLE_WORKSPACE_SERVICE_ACCOUNT_FILE")); path != "" {
		return os.ReadFile(path)
	}
	return nil, fmt.Errorf("workspace service account credentials not configured")
}

func doGoogleJSONRequest(ctx context.Context, client *http.Client, endpoint string, dest interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var body map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		return fmt.Errorf("google api request failed: %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

func findDirectoryCustomSchemaFields(schemas map[string]map[string]interface{}) (department, title, organization string) {
	for _, fields := range schemas {
		for key, raw := range fields {
			value := strings.TrimSpace(fmt.Sprint(raw))
			if value == "" {
				continue
			}
			lowerKey := strings.ToLower(strings.TrimSpace(key))
			switch {
			case department == "" && (strings.Contains(lowerKey, "department") || strings.Contains(lowerKey, "dept") || strings.Contains(lowerKey, "unit") || strings.Contains(lowerKey, "部門") || strings.Contains(lowerKey, "單位")):
				department = value
			case title == "" && (strings.Contains(lowerKey, "title") || strings.Contains(lowerKey, "job") || strings.Contains(lowerKey, "職稱")):
				title = value
			case organization == "" && (strings.Contains(lowerKey, "organization") || strings.Contains(lowerKey, "org") || strings.Contains(lowerKey, "company") || strings.Contains(lowerKey, "機構")):
				organization = value
			}
		}
	}
	return department, title, organization
}

func selectDirectoryOrganization(orgs []struct {
	Name       string `json:"name"`
	Department string `json:"department"`
	Title      string `json:"title"`
	Primary    bool   `json:"primary"`
	Type       string `json:"type"`
}) (department, title, name string) {
	for _, org := range orgs {
		if org.Primary {
			return strings.TrimSpace(org.Department), strings.TrimSpace(org.Title), strings.TrimSpace(org.Name)
		}
	}
	for _, org := range orgs {
		if strings.EqualFold(strings.TrimSpace(org.Type), "work") {
			return strings.TrimSpace(org.Department), strings.TrimSpace(org.Title), strings.TrimSpace(org.Name)
		}
	}
	if len(orgs) > 0 {
		return strings.TrimSpace(orgs[0].Department), strings.TrimSpace(orgs[0].Title), strings.TrimSpace(orgs[0].Name)
	}
	return "", "", ""
}

func selectPeopleOrganization(orgs []struct {
	Name       string `json:"name"`
	Department string `json:"department"`
	Title      string `json:"title"`
	Current    bool   `json:"current"`
	Type       string `json:"type"`
	Domain     string `json:"domain"`
}) (department, title, name string) {
	for _, org := range orgs {
		if org.Current {
			return strings.TrimSpace(org.Department), strings.TrimSpace(org.Title), strings.TrimSpace(org.Name)
		}
	}
	for _, org := range orgs {
		if strings.EqualFold(strings.TrimSpace(org.Type), "work") {
			return strings.TrimSpace(org.Department), strings.TrimSpace(org.Title), strings.TrimSpace(org.Name)
		}
	}
	if len(orgs) > 0 {
		return strings.TrimSpace(orgs[0].Department), strings.TrimSpace(orgs[0].Title), strings.TrimSpace(orgs[0].Name)
	}
	return "", "", ""
}
