package handlers

import (
	"context"
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
