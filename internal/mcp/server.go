package mcp

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"isms-privilege/internal/db"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const protocolVersion = "2025-06-18"

type Server struct {
	DB *db.DB
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string {
	return e.Message
}

type tool struct {
	Name        string                 `json:"name"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type callToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func NewServer(database *db.DB) *Server {
	return &Server{DB: database}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeRPCError(w, nil, http.StatusUnauthorized, -32001, "unauthorized")
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeRPCError(w, nil, http.StatusMethodNotAllowed, -32600, "method not allowed")
		return
	}

	var req rpcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRPCError(w, nil, http.StatusBadRequest, -32700, "parse error")
		return
	}
	if req.JSONRPC != "2.0" {
		writeRPCError(w, req.ID, http.StatusBadRequest, -32600, "invalid request")
		return
	}

	resp := s.HandleRPC(req)
	if resp == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) authorized(r *http.Request) bool {
	token := strings.TrimSpace(os.Getenv("MCP_API_TOKEN"))
	if token == "" {
		return true
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	return auth == "Bearer "+token
}

func (s *Server) HandleRPC(req rpcRequest) *rpcResponse {
	result, err := s.handle(req)
	if req.ID == nil {
		return nil
	}
	if err != nil {
		var rpcErr *rpcError
		if errors.As(err, &rpcErr) {
			return &rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: rpcErr}
		}
		return &rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: -32603, Message: err.Error()}}
	}
	return &rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: result}
}

func (s *Server) ServeStdio(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	writer := bufio.NewWriter(out)
	defer writer.Flush()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var req rpcRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			if err := writeStdioResponse(writer, rpcResponse{JSONRPC: "2.0", Error: &rpcError{Code: -32700, Message: "parse error"}}); err != nil {
				return err
			}
			continue
		}
		if req.JSONRPC != "2.0" {
			if req.ID != nil {
				if err := writeStdioResponse(writer, rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: -32600, Message: "invalid request"}}); err != nil {
					return err
				}
			}
			continue
		}

		if resp := s.HandleRPC(req); resp != nil {
			if err := writeStdioResponse(writer, *resp); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

func writeStdioResponse(writer *bufio.Writer, resp rpcResponse) error {
	if err := json.NewEncoder(writer).Encode(resp); err != nil {
		return err
	}
	return writer.Flush()
}

func (s *Server) handle(req rpcRequest) (interface{}, error) {
	switch req.Method {
	case "initialize":
		return map[string]interface{}{
			"protocolVersion": protocolVersion,
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{"listChanged": false},
			},
			"serverInfo": map[string]string{
				"name":    "aisms",
				"version": "1.0.0",
			},
		}, nil
	case "notifications/initialized":
		return nil, nil
	case "tools/list":
		return map[string]interface{}{"tools": s.tools()}, nil
	case "tools/call":
		var params callToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, &rpcError{Code: -32602, Message: "invalid params"}
		}
		return s.callTool(params)
	default:
		return nil, &rpcError{Code: -32601, Message: "method not found"}
	}
}

func (s *Server) tools() []tool {
	return []tool{
		{
			Name:        "list_dashboard_forms",
			Title:       "List Dashboard Forms",
			Description: "列出 ISMS 首頁儀表板已設定的表單。",
			InputSchema: objectSchema(nil, nil),
		},
		{
			Name:        "list_firewall_requests",
			Title:       "List Firewall Requests",
			Description: "列出 04-042 防火牆申請，可用 creator 篩選建立者。",
			InputSchema: objectSchema(map[string]interface{}{
				"creator": map[string]interface{}{"type": "string", "description": "建立者 email，可留空列出全部。"},
			}, nil),
		},
		{
			Name:        "get_firewall_request",
			Title:       "Get Firewall Request",
			Description: "依 ID 取得單筆 04-042 防火牆申請。",
			InputSchema: objectSchema(map[string]interface{}{
				"id":      map[string]interface{}{"type": "integer", "description": "防火牆申請 ID。"},
				"creator": map[string]interface{}{"type": "string", "description": "建立者 email，可留空不篩選。"},
			}, []string{"id"}),
		},
		{
			Name:        "list_asset_inventory_records",
			Title:       "List Asset Inventory Records",
			Description: "列出 04-008 資訊資產清冊，可用 keyword、status、asset_type、creator 篩選。",
			InputSchema: objectSchema(map[string]interface{}{
				"keyword":    map[string]interface{}{"type": "string", "description": "關鍵字，搜尋系統名稱、環境、資產編號、資產名稱、部門或位置。"},
				"status":     map[string]interface{}{"type": "string", "description": "狀態，例如 active、pending、closed；可留空或 all。"},
				"asset_type": map[string]interface{}{"type": "string", "description": "資產類型；可留空或 all。"},
				"creator":    map[string]interface{}{"type": "string", "description": "建立者 email，可留空列出全部。"},
			}, nil),
		},
		{
			Name:        "get_asset_inventory_record",
			Title:       "Get Asset Inventory Record",
			Description: "依 ID 取得單筆 04-008 資訊資產清冊資料。",
			InputSchema: objectSchema(map[string]interface{}{
				"id":      map[string]interface{}{"type": "integer", "description": "資訊資產清冊 ID。"},
				"creator": map[string]interface{}{"type": "string", "description": "建立者 email，可留空不篩選。"},
			}, []string{"id"}),
		},
		{
			Name:        "list_application_change_requests",
			Title:       "List Application Change Requests",
			Description: "列出 04-052 功能需求更新建議，可用 creator 篩選建立者。",
			InputSchema: objectSchema(map[string]interface{}{
				"creator": map[string]interface{}{"type": "string", "description": "建立者 email，可留空列出全部。"},
			}, nil),
		},
		{
			Name:        "get_application_change_request",
			Title:       "Get Application Change Request",
			Description: "依 ID 取得單筆 04-052 功能需求更新建議。",
			InputSchema: objectSchema(map[string]interface{}{
				"id":      map[string]interface{}{"type": "integer", "description": "功能需求更新建議 ID。"},
				"creator": map[string]interface{}{"type": "string", "description": "建立者 email，可留空不篩選。"},
			}, []string{"id"}),
		},
		{
			Name:        "list_system_platform_requests",
			Title:       "List System Platform Requests",
			Description: "列出 04-078 系統平台申請，可用 creator 篩選建立者。",
			InputSchema: objectSchema(map[string]interface{}{
				"creator": map[string]interface{}{"type": "string", "description": "建立者 email，可留空列出全部。"},
			}, nil),
		},
		{
			Name:        "get_system_platform_request",
			Title:       "Get System Platform Request",
			Description: "依 ID 取得單筆 04-078 系統平台申請。",
			InputSchema: objectSchema(map[string]interface{}{
				"id":      map[string]interface{}{"type": "integer", "description": "系統平台申請 ID。"},
				"creator": map[string]interface{}{"type": "string", "description": "建立者 email，可留空不篩選。"},
			}, []string{"id"}),
		},
	}
}

func objectSchema(properties map[string]interface{}, required []string) map[string]interface{} {
	if properties == nil {
		properties = map[string]interface{}{}
	}
	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func (s *Server) callTool(params callToolParams) (interface{}, error) {
	switch params.Name {
	case "list_dashboard_forms":
		forms, err := s.DB.ListDashboardForms()
		if err != nil {
			return nil, err
		}
		return toolJSONContent(forms)
	case "list_firewall_requests":
		rows, err := s.DB.ListFirewallRequestsByCreator(stringArg(params.Arguments, "creator"))
		if err != nil {
			return nil, err
		}
		return toolJSONContent(rows)
	case "get_firewall_request":
		id, ok := intArg(params.Arguments, "id")
		if !ok {
			return nil, &rpcError{Code: -32602, Message: "id is required"}
		}
		row, err := s.DB.GetFirewallRequestByCreator(id, stringArg(params.Arguments, "creator"))
		if err != nil {
			return nil, err
		}
		return toolJSONContent(row)
	case "list_asset_inventory_records":
		rows, err := s.DB.ListAssetInventoryRecordsByCreator(
			stringArg(params.Arguments, "keyword"),
			stringArg(params.Arguments, "status"),
			stringArg(params.Arguments, "asset_type"),
			stringArg(params.Arguments, "creator"),
		)
		if err != nil {
			return nil, err
		}
		return toolJSONContent(rows)
	case "get_asset_inventory_record":
		id, ok := intArg(params.Arguments, "id")
		if !ok {
			return nil, &rpcError{Code: -32602, Message: "id is required"}
		}
		row, err := s.DB.GetAssetInventoryRecordByCreator(id, stringArg(params.Arguments, "creator"))
		if err != nil {
			return nil, err
		}
		return toolJSONContent(row)
	case "list_application_change_requests":
		rows, err := s.DB.ListApplicationChangeRequestsByCreator(stringArg(params.Arguments, "creator"))
		if err != nil {
			return nil, err
		}
		return toolJSONContent(rows)
	case "get_application_change_request":
		id, ok := intArg(params.Arguments, "id")
		if !ok {
			return nil, &rpcError{Code: -32602, Message: "id is required"}
		}
		row, err := s.DB.GetApplicationChangeRequestByCreator(id, stringArg(params.Arguments, "creator"))
		if err != nil {
			return nil, err
		}
		return toolJSONContent(row)
	case "list_system_platform_requests":
		rows, err := s.DB.ListSystemPlatformRequestsByCreator(stringArg(params.Arguments, "creator"))
		if err != nil {
			return nil, err
		}
		return toolJSONContent(rows)
	case "get_system_platform_request":
		id, ok := intArg(params.Arguments, "id")
		if !ok {
			return nil, &rpcError{Code: -32602, Message: "id is required"}
		}
		row, err := s.DB.GetSystemPlatformRequestByCreator(id, stringArg(params.Arguments, "creator"))
		if err != nil {
			return nil, err
		}
		return toolJSONContent(row)
	default:
		return nil, &rpcError{Code: -32602, Message: fmt.Sprintf("unknown tool: %s", params.Name)}
	}
}

func toolJSONContent(v interface{}) (interface{}, error) {
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"content": []map[string]string{
			{"type": "text", "text": string(body)},
		},
		"isError": false,
	}, nil
}

func stringArg(args map[string]interface{}, name string) string {
	if args == nil {
		return ""
	}
	value, _ := args[name].(string)
	return strings.TrimSpace(value)
}

func intArg(args map[string]interface{}, name string) (int, bool) {
	if args == nil {
		return 0, false
	}
	switch value := args[name].(type) {
	case float64:
		return int(value), true
	case int:
		return value, true
	case string:
		id, err := strconv.Atoi(strings.TrimSpace(value))
		return id, err == nil
	default:
		return 0, false
	}
}

func writeRPCError(w http.ResponseWriter, id json.RawMessage, status, code int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: message},
	})
}
