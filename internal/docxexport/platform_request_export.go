package docxexport

import (
	"fmt"
	"isms-privilege/internal/models"
	"strings"
)

type PlatformRequestExportOptions struct {
	TemplatePath string
	FormName     string
	FormCode     string
	Version      string
	RecordCode   string
	Department   string
	HandlerName  string
	ManagerName  string
	ReviewNotes  string
	Request      models.SystemPlatformRequest
}

func GeneratePlatformRequest(opts PlatformRequestExportOptions) ([]byte, error) {
	if strings.TrimSpace(opts.TemplatePath) == "" {
		return nil, fmt.Errorf("template path is required")
	}

	headerXML := buildPlatformRequestHeaderXML(opts)
	documentXML := buildPlatformRequestDocumentXML(opts)
	return generateFromTemplate(opts.TemplatePath, headerXML, documentXML)
}

func buildPlatformRequestHeaderXML(opts PlatformRequestExportOptions) string {
	formName := defaultString(opts.FormName, "系統平台申請表")
	formCode := defaultString(opts.FormCode, "ISMS-04-078")
	version := defaultString(opts.Version, "1.0")
	recordCode := strings.TrimSpace(opts.RecordCode)
	department := defaultString(opts.Department, "系統科")

	return buildHeaderXML(ExportOptions{
		FormName:   formName,
		FormCode:   formCode,
		Version:    version,
		RecordCode: recordCode,
		Department: department,
	})
}

func buildPlatformRequestDocumentXML(opts PlatformRequestExportOptions) string {
	req := opts.Request
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<w:document xmlns:wpc="http://schemas.microsoft.com/office/word/2010/wordprocessingCanvas" xmlns:cx="http://schemas.microsoft.com/office/drawing/2014/chartex" xmlns:cx1="http://schemas.microsoft.com/office/drawing/2015/9/8/chartex" xmlns:cx2="http://schemas.microsoft.com/office/drawing/2015/10/21/chartex" xmlns:cx3="http://schemas.microsoft.com/office/drawing/2016/5/9/chartex" xmlns:cx4="http://schemas.microsoft.com/office/drawing/2016/5/10/chartex" xmlns:cx5="http://schemas.microsoft.com/office/drawing/2016/5/11/chartex" xmlns:cx6="http://schemas.microsoft.com/office/drawing/2016/5/12/chartex" xmlns:cx7="http://schemas.microsoft.com/office/drawing/2016/5/13/chartex" xmlns:cx8="http://schemas.microsoft.com/office/drawing/2016/5/14/chartex" xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006" xmlns:aink="http://schemas.microsoft.com/office/drawing/2016/ink" xmlns:am3d="http://schemas.microsoft.com/office/drawing/2017/model3d" xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:oel="http://schemas.microsoft.com/office/2019/extlst" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:m="http://schemas.openxmlformats.org/officeDocument/2006/math" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:wp14="http://schemas.microsoft.com/office/word/2010/wordprocessingDrawing" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:w10="urn:schemas-microsoft-com:office:word" xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:w14="http://schemas.microsoft.com/office/word/2010/wordml" xmlns:w15="http://schemas.microsoft.com/office/word/2012/wordml" xmlns:w16cei="http://schemas.microsoft.com/office/word/2026/wordml/cei" xmlns:w16cex="http://schemas.microsoft.com/office/word/2018/wordml/cex" xmlns:w16cid="http://schemas.microsoft.com/office/word/2016/wordml/cid" xmlns:w16="http://schemas.microsoft.com/office/word/2018/wordml" xmlns:w16du="http://schemas.microsoft.com/office/word/2023/wordml/word16du" xmlns:w16sdtdh="http://schemas.microsoft.com/office/word/2020/wordml/sdtdatahash" xmlns:w16sdtfl="http://schemas.microsoft.com/office/word/2024/wordml/sdtformatlock" xmlns:w16se="http://schemas.microsoft.com/office/word/2015/wordml/symex" xmlns:wpg="http://schemas.microsoft.com/office/word/2010/wordprocessingGroup" xmlns:wpi="http://schemas.microsoft.com/office/word/2010/wordprocessingInk" xmlns:wne="http://schemas.microsoft.com/office/word/2006/wordml" xmlns:wps="http://schemas.microsoft.com/office/word/2010/wordprocessingShape" mc:Ignorable="w14 w15 w16se w16cid w16 w16cex w16sdtdh w16sdtfl w16du wp14"><w:body>` +
		buildPlatformRequestTitleXML() +
		buildPlatformRequestTableXML(req) +
		buildPlatformRequestNotesXML() +
		buildPlatformRequestResultXML(opts) +
		buildPlatformRequestFooterNotesXML() +
		`<w:sectPr w:rsidR="00A948ED" w:rsidSect="00057C3B"><w:headerReference w:type="default" r:id="rId9"/><w:footerReference w:type="default" r:id="rId10"/><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="850" w:right="850" w:bottom="850" w:left="850" w:header="708" w:footer="425" w:gutter="0"/><w:cols w:space="425"/><w:docGrid w:type="lines" w:linePitch="312"/></w:sectPr></w:body></w:document>`
}

func buildPlatformRequestTitleXML() string {
	return paragraphXML("系統平台申請表", "center", true, 32) +
		paragraphXML("", "left", false, 22)
}

func buildPlatformRequestTableXML(req models.SystemPlatformRequest) string {
	return `<w:tbl><w:tblPr><w:tblW w:w="5000" w:type="pct"/><w:tblBorders><w:top w:val="single" w:sz="8" w:space="0" w:color="auto"/><w:left w:val="single" w:sz="8" w:space="0" w:color="auto"/><w:bottom w:val="single" w:sz="8" w:space="0" w:color="auto"/><w:right w:val="single" w:sz="8" w:space="0" w:color="auto"/><w:insideH w:val="single" w:sz="6" w:space="0" w:color="auto"/><w:insideV w:val="single" w:sz="6" w:space="0" w:color="auto"/></w:tblBorders><w:tblCellMar><w:top w:w="80" w:type="dxa"/><w:left w:w="80" w:type="dxa"/><w:bottom w:w="80" w:type="dxa"/><w:right w:w="80" w:type="dxa"/></w:tblCellMar></w:tblPr><w:tblGrid><w:gridCol w:w="1800"/><w:gridCol w:w="2600"/><w:gridCol w:w="1800"/><w:gridCol w:w="2600"/><w:gridCol w:w="1800"/><w:gridCol w:w="2600"/></w:tblGrid>` +
		tableRowXML(
			labelCell("填表日期", 1),
			valueCell(formatROCDate(req.RequestDate), 5),
		) +
		tableRowXML(
			labelCell("申請人", 1),
			valueCell(joinLines(
				"姓名: "+req.ApplicantName,
				"單位: "+req.ApplicantDepartment,
				"職稱: "+req.ApplicantTitle,
			), 2),
			labelCell("辦公室電話", 1),
			valueCell(req.OfficePhone, 2),
		) +
		tableRowXML(
			labelCell("電子郵件 / PI", 1),
			valueCell(joinLines("電子郵件: "+req.Email, "PI: "+req.PIName), 5),
		) +
		tableRowXML(
			labelCell("應用系統", 1),
			valueCell("應用系統名稱: "+req.SystemName, 3),
			labelCell("英文簡稱", 1),
			valueCell(req.SystemAlias, 1),
		) +
		tableRowXML(
			labelCell("系統用途說明", 1),
			valueCell(req.SystemPurpose, 5),
		) +
		tableRowXML(
			labelCell("預估使用人數", 1),
			valueCell(req.EstimatedUsers, 1),
			labelCell("限院內使用", 1),
			valueCell(req.InternalOnly, 1),
			labelCell("IP位址限制", 1),
			valueCell(req.IPRestriction, 1),
		) +
		tableRowXML(
			labelCell("申請期間", 1),
			valueCell(strings.TrimSpace(req.RequestStartDate+" 至 "+req.RequestEndDate), 5),
		) +
		tableRowXML(
			labelCell("申請樣態", 1),
			valueCell(buildRequestTypeText(req), 5),
		) +
		tableRowXML(
			labelCell("環境類別", 1),
			valueCell(req.EnvironmentType, 5),
		) +
		tableRowXML(
			labelCell("作業系統", 1),
			valueCell(buildOperatingSystemText(req), 5),
		) +
		tableRowXML(
			labelCell("硬碟容量", 1),
			valueCell(req.DiskSize, 5),
		) +
		tableRowXML(
			labelCell("特殊需求", 1),
			valueCell(req.SpecialRequirements, 5),
		) +
		tableRowXML(
			labelCell("其他", 1),
			valueCell(buildOtherRequirementsText(req), 5),
		) +
		tableRowXML(
			labelCell("備份需求", 1),
			valueCell(buildBackupText(req), 5),
		) +
		tableRowXML(
			labelCell("申請人", 1),
			valueCell(req.ApplicantSignature, 2),
			labelCell("主管簽章", 1),
			valueCell(req.SupervisorSignature, 2),
		) +
		tableRowXML(
			labelCell("備註", 1),
			valueCell(req.Remarks, 5),
		) +
		`</w:tbl>`
}

func buildPlatformRequestNotesXML() string {
	lines := []string{
		"請注意:",
		"1. 應用系統需提供完整安裝手冊及使用手冊，據以進行部署。",
		"2. 為確保安全，Linux以Rocky官方版本為準，未經同意，不採用自行編譯之版本。",
		"3. 為確保安全，申請人需配合進行系統安全性更新。",
		"4. 應用系統需先申請測試環境，俟測試完成，並交付測試報告後，始得申請部署至正式環境。",
	}
	return paragraphXML(strings.Join(lines, "\n"), "left", false, 20)
}

func buildPlatformRequestResultXML(opts PlatformRequestExportOptions) string {
	return paragraphXML("-----------------------------------------------------系統科辦理結果 Result-------------------------------------------------", "center", true, 20) +
		`<w:tbl><w:tblPr><w:tblW w:w="5000" w:type="pct"/><w:tblBorders><w:top w:val="single" w:sz="8" w:space="0" w:color="auto"/><w:left w:val="single" w:sz="8" w:space="0" w:color="auto"/><w:bottom w:val="single" w:sz="8" w:space="0" w:color="auto"/><w:right w:val="single" w:sz="8" w:space="0" w:color="auto"/><w:insideH w:val="single" w:sz="6" w:space="0" w:color="auto"/><w:insideV w:val="single" w:sz="6" w:space="0" w:color="auto"/></w:tblBorders></w:tblPr><w:tblGrid><w:gridCol w:w="2200"/><w:gridCol w:w="4200"/><w:gridCol w:w="4200"/></w:tblGrid>` +
		tableRowXML(labelCell("審核", 1), valueCell("□同意  □不同意", 2)) +
		tableRowXML(labelCell("審查意見 / 安裝或移除路徑", 1), valueCell(opts.ReviewNotes, 2)) +
		tableRowXML(valueCell("承辦人："+strings.TrimSpace(opts.HandlerName), 1), valueCell("", 1), valueCell("權責主管："+strings.TrimSpace(opts.ManagerName), 1)) +
		`</w:tbl>`
}

func buildPlatformRequestFooterNotesXML() string {
	lines := []string{
		"系統平台申請表注意事項:",
		"1. 系統平台申請，一次申請最長2年，期滿可再申請。",
		"2. 為維持系統穩定度，本處提供近兩代官方維護之作業系統(Windows、Rocky Linux)及資料庫(MS SQL、MySql)主版本。",
		"3. 若有申請網域名稱之需求者，請填寫DNS網域名稱申請書，屬本院一、二級單位或供全院使用之服務，始得申請本院第1層網域名稱。(如: 服務名稱.sinica.edu.tw)",
		"4. 應用系統須先申請測試環境，俟測試完成，並交付測試報告後，始得申請部署至正式區。",
		"5. 應用系統請提供完整安裝手冊及使用手冊，本處將據以進行部署。",
		"6. 為確保系統安全，本處 Linux 套件安裝以Rocky官方提供之相容套件為準，未經同意不應以自行編譯之套件取代。",
		"7. 請配合本處依正常程序進行主機系統安全性修正檔之更新作業，確保應用系統正常運作。",
		"8. 提供硬碟容量預設僅滿足基本作業需求，擴充空間最大以300GB為原則。",
		"9. 上述注意事項屬原則性規範，若未能符合使用需求，請另洽資訊服務處。",
	}
	return paragraphXML(strings.Join(lines, "\n"), "left", false, 18)
}

func tableRowXML(cells ...string) string {
	return `<w:tr>` + strings.Join(cells, "") + `</w:tr>`
}

func labelCell(text string, span int) string {
	return tableCellXML(text, span, true)
}

func valueCell(text string, span int) string {
	return tableCellXML(text, span, false)
}

func tableCellXML(text string, span int, shaded bool) string {
	if span < 1 {
		span = 1
	}
	cell := `<w:tc><w:tcPr>`
	if span > 1 {
		cell += fmt.Sprintf(`<w:gridSpan w:val="%d"/>`, span)
	}
	if shaded {
		cell += `<w:shd w:val="clear" w:color="auto" w:fill="D9D9D9"/>`
	}
	cell += `<w:vAlign w:val="center"/></w:tcPr>`
	cell += multilineParagraphXML(text, "left", shaded)
	cell += `</w:tc>`
	return cell
}

func multilineParagraphXML(text, align string, bold bool) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	var b strings.Builder
	for idx, line := range lines {
		if idx > 0 {
			b.WriteString(`<w:p/>`)
		}
		b.WriteString(paragraphXML(line, align, bold, 20))
	}
	return b.String()
}

func paragraphXML(text, align string, bold bool, size int) string {
	if align == "" {
		align = "left"
	}
	if size <= 0 {
		size = 20
	}
	return `<w:p><w:pPr><w:jc w:val="` + align + `"/></w:pPr><w:r><w:rPr><w:rFonts w:eastAsia="標楷體" w:hint="eastAsia"/>` + boolTag(bold, "<w:b/><w:bCs/>") + fmt.Sprintf(`<w:sz w:val="%d"/><w:szCs w:val="%d"/>`, size, size) + `</w:rPr><w:t xml:space="preserve">` + xmlEscape(text) + `</w:t></w:r></w:p>`
}

func boolTag(enabled bool, value string) string {
	if enabled {
		return value
	}
	return ""
}

func joinLines(lines ...string) string {
	var filtered []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}

func formatROCDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if len(raw) == 8 && !strings.Contains(raw, "-") {
		return raw[:4] + "-" + raw[4:6] + "-" + raw[6:]
	}
	return raw
}

func buildRequestTypeText(req models.SystemPlatformRequest) string {
	base := req.RequestType
	if strings.TrimSpace(base) == "關機保留" {
		return strings.TrimSpace(base + " " + req.ShutdownRetainMonths + " 個月，理由: " + req.ShutdownReason)
	}
	if strings.TrimSpace(req.ShutdownReason) != "" {
		return strings.TrimSpace(base + "，理由: " + req.ShutdownReason)
	}
	return base
}

func buildOperatingSystemText(req models.SystemPlatformRequest) string {
	if strings.TrimSpace(req.OperatingSystem) == "其他" && strings.TrimSpace(req.OperatingSystemOther) != "" {
		return "其他: " + req.OperatingSystemOther
	}
	if strings.TrimSpace(req.OperatingSystemOther) != "" {
		return req.OperatingSystem + " / " + req.OperatingSystemOther
	}
	return req.OperatingSystem
}

func buildOtherRequirementsText(req models.SystemPlatformRequest) string {
	return joinLines(
		"網域名稱設定: "+req.DomainSettings,
		"其他: "+req.OtherRequirements,
	)
}

func buildBackupText(req models.SystemPlatformRequest) string {
	if strings.TrimSpace(req.BackupRequired) == "否" {
		if strings.TrimSpace(req.BackupReason) != "" {
			return "否，理由: " + req.BackupReason
		}
		return "否"
	}
	if strings.TrimSpace(req.BackupRequirements) != "" {
		return "是，特殊備份需求: " + req.BackupRequirements
	}
	return "是"
}
