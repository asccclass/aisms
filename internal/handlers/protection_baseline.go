package handlers

import (
	"bytes"
	"fmt"
	"isms-privilege/internal/datefmt"
	"isms-privilege/internal/db"
	"isms-privilege/internal/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

func (h *Handler) ListProtectionBaselineRecords(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.ListProtectionBaselineRecordsByCreator(GetUserEmail(r))
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, rows)
}

func (h *Handler) GetProtectionBaselineTemplate(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, db.ProtectionBaselineTemplateControls())
}

func (h *Handler) GetProtectionBaselineRecord(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/protection-baselines/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	row, err := h.DB.GetProtectionBaselineRecordByCreator(id, GetUserEmail(r))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, 200, row)
}

func (h *Handler) CreateProtectionBaselineRecord(w http.ResponseWriter, r *http.Request) {
	var req models.ProtectionBaselineRecord
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	req.Creator = GetUserEmail(r)
	if strings.TrimSpace(req.Status) == "" {
		req.Status = "active"
	}
	id, err := h.DB.CreateProtectionBaselineRecord(&req)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	record, err := h.DB.GetProtectionBaselineRecordByCreator(int(id), GetUserEmail(r))
	if err != nil {
		writeJSON(w, 201, map[string]int64{"id": id})
		return
	}
	writeJSON(w, 201, record)
}

func (h *Handler) UpdateProtectionBaselineRecord(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/protection-baselines/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	var req models.ProtectionBaselineRecord
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	req.ID = id
	existing, _ := h.DB.GetProtectionBaselineRecordByCreator(id, GetUserEmail(r))
	if existing == nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	req.Creator = existing.Creator
	if err := h.DB.UpdateProtectionBaselineRecord(&req); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, req)
}

func (h *Handler) DeleteProtectionBaselineRecord(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/protection-baselines/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	existing, _ := h.DB.GetProtectionBaselineRecordByCreator(id, GetUserEmail(r))
	if existing == nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	if err := h.DB.DeleteProtectionBaselineRecordByCreator(id, GetUserEmail(r)); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"message": "deleted"})
}

func (h *Handler) ExportProtectionBaselineXLSX(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/protection-baselines/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	record, err := h.DB.GetProtectionBaselineRecordByCreator(id, GetUserEmail(r))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}

	f := excelize.NewFile()
	sheetName := "ISMS-04-069資通系統防護基準執行說明表"
	f.SetSheetName("Sheet1", sheetName)

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"D9EAF7"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	cellStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "top", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})
	centerCellStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	colWidths := map[string]float64{
		"A": 8, "B": 14, "C": 18, "D": 10, "E": 40,
		"F": 14, "G": 28, "H": 18, "I": 28, "J": 18, "K": 4,
	}
	for col, width := range colWidths {
		_ = f.SetColWidth(sheetName, col, col, width)
	}

	_ = f.MergeCell(sheetName, "A1", "E1")
	_ = f.MergeCell(sheetName, "F1", "K1")
	_ = f.SetCellValue(sheetName, "A1", "表單或紀錄名稱")
	_ = f.SetCellValue(sheetName, "F1", "表單或紀錄編號")
	_ = f.SetCellValue(sheetName, "C1", "ISMS-04-069資通系統防護基準執行說明表")
	_ = f.SetCellValue(sheetName, "A2", "機密等級")
	_ = f.SetCellValue(sheetName, "C2", "內部使用")
	_ = f.SetCellValue(sheetName, "D2", "版本")
	_ = f.SetCellValue(sheetName, "E2", "1.3")
	_ = f.SetCellValue(sheetName, "F2", "權責單位")
	_ = f.SetCellValue(sheetName, "G2", "發展科")
	_ = f.MergeCell(sheetName, "A3", "J3")
	_ = f.SetCellValue(sheetName, "A3", "資通系統防護基準執行說明表")
	_ = f.MergeCell(sheetName, "A4", "J4")
	_ = f.SetCellValue(sheetName, "A4", "因應資安法施行-資通系統符合資安法規之安全強化說明")

	introLines := []string{
		"● 資通安全管理法(以下簡稱資安法)於107年5月11日立法院完成三讀，107年6月6日由總統公布，以加速建構國家資通安全環境，保障國家安全",
		"● 依據資安法第7條第1項，訂定「資通安全責任等級分級辦法」，並於107年11月21日公告，另於115年1月7日修正，該辦法明定各機關應依附表之規定，辦理其資通安全責任等級應辦事項",
		"● 本報告說明「資通安全責任等級分級辦法」中，資通系統防護基準要求之重要控制措施，協助機關了解技術面實作項目，以強化系統安全性，並符合法規要求",
		"依「資訊系統分級與資安防護基準作業規定」，包括7個構面、27項控制措施類別",
		"•系統安全等級為普；僅需篩選普的項目，進行評估  (共45項)",
		"•系統安全等級為中；需篩選普的項目和中的項目，進行評估  (共62項)",
		"•系統安全等級為高；全選，需全部評估(普,中,高項目)  (共80項)",
	}
	row := 5
	for _, line := range introLines {
		_ = f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("J%d", row))
		_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), line)
		row++
	}
	row++

	_ = f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "系統名稱："+record.SystemName)
	row++
	_ = f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "系統安全等級："+record.SecurityLevel)
	row++
	_ = f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "填寫人："+record.FilledBy)
	row++
	_ = f.MergeCell(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row))
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "填表日期："+datefmt.NormalizeDate(record.FormDate))
	row++

	headers := []string{
		"項次", "構面", "措施內容", "系統防護需求等級", "措施說明",
		"是否套用?\n(Y、N、不適用)", "執行說明", "執行結果符合性\n(符合、不符合、不適用)", "發現缺失及建議改善事項", "備註",
	}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, row)
		_ = f.SetCellValue(sheetName, cell, header)
		_ = f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	_ = f.SetRowHeight(sheetName, row, 36)

	dataStartRow := row + 1
	for idx, control := range record.Controls {
		currentRow := dataStartRow + idx
		values := []interface{}{
			control.ItemNo,
			control.DomainName,
			control.ControlCategory,
			control.RequirementLevel,
			control.ControlDescription,
			control.Applies,
			control.ImplementationNotes,
			control.Compliance,
			control.Finding,
			control.Remarks,
		}
		for i, value := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, currentRow)
			_ = f.SetCellValue(sheetName, cell, value)
			style := cellStyle
			if i == 0 || i == 3 || i == 5 || i == 7 {
				style = centerCellStyle
			}
			_ = f.SetCellStyle(sheetName, cell, cell, style)
		}
		_ = f.SetRowHeight(sheetName, currentRow, 42)
	}

	reviewRow := dataStartRow + len(record.Controls)
	_ = f.MergeCell(sheetName, fmt.Sprintf("A%d", reviewRow), fmt.Sprintf("D%d", reviewRow))
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", reviewRow), "審查人："+record.Reviewer)
	_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", reviewRow), fmt.Sprintf("D%d", reviewRow), cellStyle)
	reviewDateRow := reviewRow + 1
	_ = f.MergeCell(sheetName, fmt.Sprintf("A%d", reviewDateRow), fmt.Sprintf("D%d", reviewDateRow))
	_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", reviewDateRow), "審查日期："+datefmt.NormalizeDate(record.ReviewDate))
	_ = f.SetCellStyle(sheetName, fmt.Sprintf("A%d", reviewDateRow), fmt.Sprintf("D%d", reviewDateRow), cellStyle)

	_ = f.SetCellStyle(sheetName, "A1", "K2", titleStyle)
	_ = f.SetCellStyle(sheetName, "A3", fmt.Sprintf("J%d", reviewDateRow), cellStyle)
	_ = f.SetPanes(sheetName, &excelize.Panes{Freeze: true, YSplit: dataStartRow - 1, TopLeftCell: fmt.Sprintf("A%d", dataStartRow), ActivePane: "bottomLeft"})

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	datePart := datefmt.NormalizeDate(record.FormDate)
	if datePart == "" {
		datePart = datefmt.Today()
	}
	filename := fmt.Sprintf("ISMS-04-069_%s_%s.xlsx", sanitizeFilenamePart(firstNonEmpty(record.SystemName, "資通系統防護基準執行說明表")), sanitizeFilenamePart(datePart))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", filename))
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	_, _ = w.Write(buf.Bytes())
}
