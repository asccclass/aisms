package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"isms-privilege/internal/db"
	"isms-privilege/internal/models"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestMCPInitializeAndToolsList(t *testing.T) {
	server := newTestServer(t)

	initResp := postRPC(t, server, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`)
	if initResp.Code != http.StatusOK {
		t.Fatalf("initialize status = %d, want %d", initResp.Code, http.StatusOK)
	}
	var initBody map[string]interface{}
	decodeBody(t, initResp, &initBody)
	result := initBody["result"].(map[string]interface{})
	if result["protocolVersion"] != protocolVersion {
		t.Fatalf("protocolVersion = %v, want %s", result["protocolVersion"], protocolVersion)
	}

	listResp := postRPC(t, server, `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`)
	if listResp.Code != http.StatusOK {
		t.Fatalf("tools/list status = %d, want %d", listResp.Code, http.StatusOK)
	}
	var listBody map[string]interface{}
	decodeBody(t, listResp, &listBody)
	tools := listBody["result"].(map[string]interface{})["tools"].([]interface{})
	if len(tools) != 9 {
		t.Fatalf("tools/list returned %d tools, want 9", len(tools))
	}
}

func TestMCPCallListFirewallRequests(t *testing.T) {
	server := newTestServer(t)

	resp := postRPC(t, server, `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_firewall_requests","arguments":{}}}`)
	if resp.Code != http.StatusOK {
		t.Fatalf("tools/call status = %d, want %d", resp.Code, http.StatusOK)
	}

	var body map[string]interface{}
	decodeBody(t, resp, &body)
	content := body["result"].(map[string]interface{})["content"].([]interface{})
	if len(content) != 1 {
		t.Fatalf("content length = %d, want 1", len(content))
	}
	text := content[0].(map[string]interface{})["text"].(string)
	if text == "" || text == "null" {
		t.Fatalf("tool response text should contain JSON data")
	}
}

func TestMCPRequiresBearerTokenWhenConfigured(t *testing.T) {
	t.Setenv("MCP_API_TOKEN", "secret")
	server := newTestServer(t)

	resp := postRPC(t, server, `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`)
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized tools/list status = %d, want %d", resp.Code, http.StatusUnauthorized)
	}

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`))
	req.Header.Set("Authorization", "Bearer secret")
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("authorized tools/list status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestMCPServeStdioHandlesJSONRPCLines(t *testing.T) {
	server := newTestServer(t)
	input := bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}` + "\n")
	var output bytes.Buffer

	if err := server.ServeStdio(input, &output); err != nil {
		t.Fatalf("ServeStdio failed: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(output.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal failed: %v; output=%s", err, output.String())
	}
	tools := body["result"].(map[string]interface{})["tools"].([]interface{})
	if len(tools) != 9 {
		t.Fatalf("stdio tools/list returned %d tools, want 9", len(tools))
	}
}

func TestMCPAdditionalFormQueryTools(t *testing.T) {
	server := newTestServer(t)
	ids := seedMCPQueryToolRecords(t, server.DB)

	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "list asset inventory records",
			body: `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_asset_inventory_records","arguments":{"keyword":"資產查詢","status":"active","asset_type":"server","creator":"owner@example.com"}}}`,
			want: "資產查詢系統",
		},
		{
			name: "get asset inventory record",
			body: fmt.Sprintf(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_asset_inventory_record","arguments":{"id":%d,"creator":"owner@example.com"}}}`, ids.assetID),
			want: "ASSET-MCP-001",
		},
		{
			name: "list application change requests",
			body: `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_application_change_requests","arguments":{"creator":"owner@example.com"}}}`,
			want: "MCP 功能查詢",
		},
		{
			name: "get application change request",
			body: fmt.Sprintf(`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_application_change_request","arguments":{"id":%d,"creator":"owner@example.com"}}}`, ids.applicationID),
			want: "ACR-MCP-001",
		},
		{
			name: "list system platform requests",
			body: `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"list_system_platform_requests","arguments":{"creator":"owner@example.com"}}}`,
			want: "平台查詢系統",
		},
		{
			name: "get system platform request",
			body: fmt.Sprintf(`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"get_system_platform_request","arguments":{"id":%d,"creator":"owner@example.com"}}}`, ids.platformID),
			want: "10.99.0.10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := postRPC(t, server, tt.body)
			if resp.Code != http.StatusOK {
				t.Fatalf("tools/call status = %d, want %d", resp.Code, http.StatusOK)
			}
			text := toolText(t, resp)
			if !bytes.Contains([]byte(text), []byte(tt.want)) {
				t.Fatalf("tool response should contain %q; text=%s", tt.want, text)
			}
		})
	}
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	database, err := db.New(filepath.Join(t.TempDir(), "isms.db"))
	if err != nil {
		t.Fatalf("db.New failed: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return NewServer(database)
}

func postRPC(t *testing.T, server *Server, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)
	return rr
}

func decodeBody(t *testing.T, rr *httptest.ResponseRecorder, target interface{}) {
	t.Helper()
	if err := json.Unmarshal(rr.Body.Bytes(), target); err != nil {
		t.Fatalf("json.Unmarshal failed: %v; body=%s", err, rr.Body.String())
	}
}

func toolText(t *testing.T, rr *httptest.ResponseRecorder) string {
	t.Helper()
	var body map[string]interface{}
	decodeBody(t, rr, &body)
	if body["error"] != nil {
		t.Fatalf("tool call returned error: %v", body["error"])
	}
	content := body["result"].(map[string]interface{})["content"].([]interface{})
	if len(content) != 1 {
		t.Fatalf("content length = %d, want 1", len(content))
	}
	return content[0].(map[string]interface{})["text"].(string)
}

type mcpQueryToolRecordIDs struct {
	assetID       int64
	applicationID int64
	platformID    int64
}

func seedMCPQueryToolRecords(t *testing.T, database *db.DB) mcpQueryToolRecordIDs {
	t.Helper()
	assetID, err := database.CreateAssetInventoryRecord(&models.AssetInventoryRecord{
		SystemName:        "資產查詢系統",
		Environment:       "正式",
		AssetCode:         "ASSET-MCP-001",
		AssetType:         "server",
		AssetName:         "MCP Asset Server",
		ManagerDepartment: "資訊服務處",
		Status:            "active",
		Creator:           "owner@example.com",
	})
	if err != nil {
		t.Fatalf("CreateAssetInventoryRecord failed: %v", err)
	}
	applicationID, err := database.CreateApplicationChangeRequest(&models.ApplicationChangeRequest{
		Suggestor:       "王小明",
		FormDate:        "2026-10-06",
		RecordNumber:    "ACR-MCP-001",
		SystemName:      "功能查詢系統",
		FeatureName:     "MCP 功能查詢",
		MeetingDecision: "同意",
		Status:          "active",
		Creator:         "owner@example.com",
	})
	if err != nil {
		t.Fatalf("CreateApplicationChangeRequest failed: %v", err)
	}
	platformID, err := database.CreateSystemPlatformRequest(&models.SystemPlatformRequest{
		RequestDate:     "2026-10-06",
		ApplicantName:   "王小明",
		SystemName:      "平台查詢系統",
		AssignedIP:      "10.99.0.10",
		EnvironmentType: "正式環境",
		Status:          "active",
		Creator:         "owner@example.com",
	})
	if err != nil {
		t.Fatalf("CreateSystemPlatformRequest failed: %v", err)
	}
	return mcpQueryToolRecordIDs{
		assetID:       assetID,
		applicationID: applicationID,
		platformID:    platformID,
	}
}
