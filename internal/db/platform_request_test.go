package db

import (
	"isms-privilege/internal/models"
	"testing"
)

func TestSystemPlatformRequestAssignedIPCRUD(t *testing.T) {
	d, err := New(":memory:")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer d.Close()

	record := &models.SystemPlatformRequest{
		RequestDate:   "2026-09-16",
		ApplicantName: "王小明",
		SystemName:    "病歷平台",
		IPRestriction: "10.10.1.0/24",
		AssignedIP:    "10.10.1.5",
		Status:        "active",
		Creator:       "owner@example.com",
	}
	id, err := d.CreateSystemPlatformRequest(record)
	if err != nil {
		t.Fatalf("CreateSystemPlatformRequest() error = %v", err)
	}

	got, err := d.GetSystemPlatformRequestByCreator(int(id), "owner@example.com")
	if err != nil {
		t.Fatalf("GetSystemPlatformRequestByCreator() error = %v", err)
	}
	if got.AssignedIP != "10.10.1.5" {
		t.Fatalf("AssignedIP = %q, want 10.10.1.5", got.AssignedIP)
	}

	got.AssignedIP = "10.10.1.6"
	if err := d.UpdateSystemPlatformRequest(got); err != nil {
		t.Fatalf("UpdateSystemPlatformRequest() error = %v", err)
	}
	updated, err := d.GetSystemPlatformRequestByCreator(int(id), "owner@example.com")
	if err != nil {
		t.Fatalf("GetSystemPlatformRequestByCreator(updated) error = %v", err)
	}
	if updated.AssignedIP != "10.10.1.6" {
		t.Fatalf("updated AssignedIP = %q, want 10.10.1.6", updated.AssignedIP)
	}
}

func TestSystemPlatformRequestsSortByAssignedIP(t *testing.T) {
	d, err := New(":memory:")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer d.Close()

	records := []models.SystemPlatformRequest{
		{ApplicantName: "王小明", SystemName: "空白IP系統", Status: "active", Creator: "owner@example.com"},
		{ApplicantName: "王小明", SystemName: "十號系統", AssignedIP: "10.0.0.10", Status: "active", Creator: "owner@example.com"},
		{ApplicantName: "王小明", SystemName: "二號系統", AssignedIP: "10.0.0.2", Status: "active", Creator: "owner@example.com"},
	}
	for i := range records {
		if _, err := d.CreateSystemPlatformRequest(&records[i]); err != nil {
			t.Fatalf("CreateSystemPlatformRequest(%s) error = %v", records[i].SystemName, err)
		}
	}

	list, err := d.ListSystemPlatformRequestsByCreator("owner@example.com")
	if err != nil {
		t.Fatalf("ListSystemPlatformRequestsByCreator() error = %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("list length = %d, want 3", len(list))
	}
	got := []string{list[0].SystemName, list[1].SystemName, list[2].SystemName}
	want := []string{"二號系統", "十號系統", "空白IP系統"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sorted names = %+v, want %+v", got, want)
		}
	}
}
