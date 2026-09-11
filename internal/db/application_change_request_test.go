package db

import (
	"isms-privilege/internal/models"
	"testing"
)

func TestApplicationChangeRequestCRUDByCreator(t *testing.T) {
	d, err := New(":memory:")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer d.Close()

	record := &models.ApplicationChangeRequest{
		Suggestor:                 "王小明",
		FormDate:                  "2026-09-11",
		SystemName:                "人事系統",
		FeatureName:               "批次匯入功能",
		IsRequiredFeature:         "是",
		IsMajorImpact:             "否",
		ExpectedOnlineDate:        "2026-10-01",
		BackgroundDescription:     "減少人工輸入時間",
		ExistingSecurityMeasures:  "否",
		SecurityScopeInvolved:     "是",
		AccessControlMeasures:     "最小權限",
		InformationServiceOpinion: "由雙方共同商議可行性",
		Status:                    "active",
		Creator:                   "owner@example.com",
	}
	id, err := d.CreateApplicationChangeRequest(record)
	if err != nil {
		t.Fatalf("CreateApplicationChangeRequest() error = %v", err)
	}
	record.ID = int(id)

	other := *record
	other.ID = 0
	other.FeatureName = "其他功能"
	other.Creator = "other@example.com"
	if _, err := d.CreateApplicationChangeRequest(&other); err != nil {
		t.Fatalf("CreateApplicationChangeRequest(other) error = %v", err)
	}

	list, err := d.ListApplicationChangeRequestsByCreator("owner@example.com")
	if err != nil {
		t.Fatalf("ListApplicationChangeRequestsByCreator() error = %v", err)
	}
	if len(list) != 1 || list[0].FeatureName != "批次匯入功能" {
		t.Fatalf("creator scoped list = %+v, want one owner record", list)
	}

	got, err := d.GetApplicationChangeRequestByCreator(record.ID, "owner@example.com")
	if err != nil {
		t.Fatalf("GetApplicationChangeRequestByCreator() error = %v", err)
	}
	if got.SystemName != record.SystemName || got.AccessControlMeasures != record.AccessControlMeasures {
		t.Fatalf("GetApplicationChangeRequestByCreator() = %+v, want %+v", got, record)
	}

	got.Status = "closed"
	got.MeetingDecision = "排入下版開發"
	if err := d.UpdateApplicationChangeRequest(got); err != nil {
		t.Fatalf("UpdateApplicationChangeRequest() error = %v", err)
	}
	updated, err := d.GetApplicationChangeRequestByCreator(record.ID, "owner@example.com")
	if err != nil {
		t.Fatalf("GetApplicationChangeRequestByCreator(updated) error = %v", err)
	}
	if updated.Status != "closed" || updated.MeetingDecision != "排入下版開發" {
		t.Fatalf("updated record = %+v, want closed with meeting decision", updated)
	}

	if err := d.DeleteApplicationChangeRequestByCreator(record.ID, "other@example.com"); err != nil {
		t.Fatalf("DeleteApplicationChangeRequestByCreator(other) error = %v", err)
	}
	if _, err := d.GetApplicationChangeRequestByCreator(record.ID, "owner@example.com"); err != nil {
		t.Fatalf("record should remain after other creator delete: %v", err)
	}

	if err := d.DeleteApplicationChangeRequestByCreator(record.ID, "owner@example.com"); err != nil {
		t.Fatalf("DeleteApplicationChangeRequestByCreator(owner) error = %v", err)
	}
	if list, err := d.ListApplicationChangeRequestsByCreator("owner@example.com"); err != nil || len(list) != 0 {
		t.Fatalf("owner list after delete = %+v, err = %v, want empty", list, err)
	}
}
