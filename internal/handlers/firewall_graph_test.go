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