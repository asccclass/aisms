package db

import (
	"isms-privilege/internal/models"
	"testing"
)

func TestAssetInventoryEnvironmentCRUD(t *testing.T) {
	d, err := New(":memory:")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer d.Close()

	record := &models.AssetInventoryRecord{
		SystemName:        "全院學習系統",
		Environment:       "測試",
		AssetCode:         "ISMS-SW-T01",
		AssetType:         "軟體類",
		AssetName:         "全院學習系統 API",
		Quantity:          "1",
		ManagerDepartment: "資訊服務處",
		Status:            "active",
		Creator:           "owner@example.com",
	}
	id, err := d.CreateAssetInventoryRecord(record)
	if err != nil {
		t.Fatalf("CreateAssetInventoryRecord() error = %v", err)
	}

	got, err := d.GetAssetInventoryRecordByCreator(int(id), "owner@example.com")
	if err != nil {
		t.Fatalf("GetAssetInventoryRecordByCreator() error = %v", err)
	}
	if got.Environment != "測試" {
		t.Fatalf("Environment = %q, want 測試", got.Environment)
	}

	got.Environment = "正式"
	if err := d.UpdateAssetInventoryRecord(got); err != nil {
		t.Fatalf("UpdateAssetInventoryRecord() error = %v", err)
	}
	updated, err := d.GetAssetInventoryRecordByCreator(int(id), "owner@example.com")
	if err != nil {
		t.Fatalf("GetAssetInventoryRecordByCreator(updated) error = %v", err)
	}
	if updated.Environment != "正式" {
		t.Fatalf("updated Environment = %q, want 正式", updated.Environment)
	}

	updated.Environment = ""
	if err := d.UpdateAssetInventoryRecord(updated); err != nil {
		t.Fatalf("UpdateAssetInventoryRecord(blank environment) error = %v", err)
	}
	blank, err := d.GetAssetInventoryRecordByCreator(int(id), "owner@example.com")
	if err != nil {
		t.Fatalf("GetAssetInventoryRecordByCreator(blank) error = %v", err)
	}
	if blank.Environment != "" {
		t.Fatalf("blank Environment = %q, want empty", blank.Environment)
	}
}
