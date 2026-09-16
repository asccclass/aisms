package handlers

import (
	"encoding/json"
	"isms-privilege/internal/db"
	"isms-privilege/internal/models"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestListFirewallRequestsForGraph(t *testing.T) {
	tempDB := filepath.Join(t.TempDir(), "test_firewall.db")
	database, err := db.New(tempDB)
	if err != nil {
		t.Fatalf("failed to create temp db: %v", err)
	}
	defer database.Close()

	h := New(database, nil)
	reqObj := models.FirewallRequest{
		SystemName:    "核心測試系統",
		SourceZone:    "外網區",
		SourceIP:      "140.109.1.50",
		DestinationIP: "192.168.10.100",
		ProtocolType:  "TCP: 443",
		FirewallID:    "FW-TEST-001",
		Status:        "active",
		Creator:       "test@example.com",
	}
	if _, err := database.CreateFirewallRequest(&reqObj); err != nil {
		t.Fatalf("CreateFirewallRequest failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/firewall-requests", nil)
	rr := httptest.NewRecorder()
	h.ListFirewallRequests(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListFirewallRequests status = %d, want %d", rr.Code, http.StatusOK)
	}
	var list []models.FirewallRequest
	if err := json.Unmarshal(rr.Body.Bytes(), &list); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if len(list) == 0 {
		t.Fatalf("ListFirewallRequests returned 0 items, want >= 1")
	}
}

func TestListFirewallRequestsUsesPlatformSystemNames(t *testing.T) {
	tempDB := filepath.Join(t.TempDir(), "test_firewall_platform_names.db")
	database, err := db.New(tempDB)
	if err != nil {
		t.Fatalf("failed to create temp db: %v", err)
	}
	defer database.Close()

	h := New(database, nil)
	sourcePlatform := models.SystemPlatformRequest{
		SystemName:    "來源平台系統",
		AssignedIP:    "10.10.1.5/32",
		IPRestriction: "10.20.0.0/24",
		Status:        "active",
	}
	if _, err := database.CreateSystemPlatformRequest(&sourcePlatform); err != nil {
		t.Fatalf("CreateSystemPlatformRequest source failed: %v", err)
	}
	destinationPlatform := models.SystemPlatformRequest{
		SystemName:    "目的平台系統",
		AssignedIP:    "172.16.8.20",
		IPRestriction: "172.16.0.0/16",
		Status:        "active",
	}
	if _, err := database.CreateSystemPlatformRequest(&destinationPlatform); err != nil {
		t.Fatalf("CreateSystemPlatformRequest destination failed: %v", err)
	}
	firewallReq := models.FirewallRequest{
		SystemName:      "原防火牆系統名稱",
		Action:          "允許",
		SourceIP:        "10.10.1.5",
		DestinationIP:   "172.16.8.20/32",
		ProtocolType:    "TCP: 443",
		RequestDate:     "2026-09-16",
		FirewallID:      "FW-TEST-078",
		Status:          "active",
		RuleDescription: "平台申請 IP 對應測試",
	}
	if _, err := database.CreateFirewallRequest(&firewallReq); err != nil {
		t.Fatalf("CreateFirewallRequest failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/firewall-requests", nil)
	rr := httptest.NewRecorder()
	h.ListFirewallRequests(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("ListFirewallRequests status = %d, want %d", rr.Code, http.StatusOK)
	}
	var list []models.FirewallRequest
	if err := json.Unmarshal(rr.Body.Bytes(), &list); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	var got *models.FirewallRequest
	for i := range list {
		if list[i].FirewallID == "FW-TEST-078" {
			got = &list[i]
			break
		}
	}
	if got == nil {
		t.Fatalf("ListFirewallRequests did not return FW-TEST-078; got %d items", len(list))
	}
	want := "來源平台系統--目的平台系統"
	if got.SystemName != want {
		t.Fatalf("SystemName = %q, want %q", got.SystemName, want)
	}
}
