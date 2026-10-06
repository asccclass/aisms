package mcp

import (
	"bytes"
	"encoding/json"
	"isms-privilege/internal/db"
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
	if len(tools) != 3 {
		t.Fatalf("tools/list returned %d tools, want 3", len(tools))
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
