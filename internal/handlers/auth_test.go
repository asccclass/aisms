package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUserEmailDecodesEscapedEmail(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), userEmailKey, "andyliu%40as.edu.tw"))

	got := GetUserEmail(req)
	want := "andyliu@as.edu.tw"
	if got != want {
		t.Fatalf("GetUserEmail() = %q, want %q", got, want)
	}
}

func TestGetUserEmailKeepsPlainEmail(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), userEmailKey, "andyliu@as.edu.tw"))

	got := GetUserEmail(req)
	want := "andyliu@as.edu.tw"
	if got != want {
		t.Fatalf("GetUserEmail() = %q, want %q", got, want)
	}
}

func TestMeReturnsDecodedEmail(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.AddCookie(&http.Cookie{Name: "admin_email", Value: "andyliu%40as.edu.tw"})
	req.AddCookie(&http.Cookie{Name: "admin_name", Value: "Andy"})
	rr := httptest.NewRecorder()

	h.Me(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Me() status = %d, want %d", rr.Code, http.StatusOK)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if got, want := body["email"], "andyliu@as.edu.tw"; got != want {
		t.Fatalf("Me() email = %v, want %q", got, want)
	}
}
