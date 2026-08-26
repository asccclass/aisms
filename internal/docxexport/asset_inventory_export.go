package docxexport

import (
	"fmt"
	"isms-privilege/internal/models"
	"strings"
)

type AssetInventoryExportOptions struct {
	TemplatePath string
	FormName     string
	FormCode     string
	Version      string
	Department   string
	Records      []models.AssetInventoryRecord
}

func GenerateAssetInventory(opts AssetInventoryExportOptions) ([]byte, error) {
	if strings.TrimSpace(opts.TemplatePath) == "" {
		return nil, fmt.Errorf("template path is required")
	}

	headerXML := buildAssetInventoryHeaderXML(opts)
	documentXML := buildAssetInventoryDocumentXML(opts)
	return generateFromTemplate(opts.TemplatePath, headerXML, documentXML)
}

func buildAssetInventoryHeaderXML(opts AssetInventoryExportOptions) string {
	return buildHeaderXML(ExportOptions{
		FormName:   defaultString(opts.FormName, "資訊資產清冊"),
		FormCode:   defaultString(opts.FormCode, "ISMS-04-008"),
		Version:    defaultString(opts.Version, "1.6"),
		Department: defaultString(opts.Department, "資安科"),
	})
}

func buildAssetInventoryDocumentXML(opts AssetInventoryExportOptions) string {
	var sections strings.Builder
	sections.WriteString(paragraphXML("資訊資產清冊", "center", true, 30))
	sections.WriteString(paragraphXML("依 ISMS-04-008 匯出", "center", false, 18))
	sections.WriteString(paragraphXML("", "left", false, 18))

	if len(opts.Records) == 0 {
		sections.WriteString(paragraphXML("目前沒有可匯出的資訊資產資料。", "left", false, 20))
	} else {
		for idx, record := range opts.Records {
			if idx > 0 {
				sections.WriteString(`<w:p><w:r><w:br w:type="page"/></w:r></w:p>`)
			}
			sections.WriteString(buildAssetInventoryRecordXML(idx+1, record))
		}
	}

	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<w:document xmlns:wpc="http://schemas.microsoft.com/office/word/2010/wordprocessingCanvas" xmlns:cx="http://schemas.microsoft.com/office/drawing/2014/chartex" xmlns:cx1="http://schemas.microsoft.com/office/drawing/2015/9/8/chartex" xmlns:cx2="http://schemas.microsoft.com/office/drawing/2015/10/21/chartex" xmlns:cx3="http://schemas.microsoft.com/office/drawing/2016/5/9/chartex" xmlns:cx4="http://schemas.microsoft.com/office/drawing/2016/5/10/chartex" xmlns:cx5="http://schemas.microsoft.com/office/drawing/2016/5/11/chartex" xmlns:cx6="http://schemas.microsoft.com/office/drawing/2016/5/12/chartex" xmlns:cx7="http://schemas.microsoft.com/office/drawing/2016/5/13/chartex" xmlns:cx8="http://schemas.microsoft.com/office/drawing/2016/5/14/chartex" xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006" xmlns:aink="http://schemas.microsoft.com/office/drawing/2016/ink" xmlns:am3d="http://schemas.microsoft.com/office/drawing/2017/model3d" xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:oel="http://schemas.microsoft.com/office/2019/extlst" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:m="http://schemas.openxmlformats.org/officeDocument/2006/math" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:wp14="http://schemas.microsoft.com/office/word/2010/wordprocessingDrawing" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:w10="urn:schemas-microsoft-com:office:word" xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:w14="http://schemas.microsoft.com/office/word/2010/wordml" xmlns:w15="http://schemas.microsoft.com/office/word/2012/wordml" xmlns:w16cei="http://schemas.microsoft.com/office/word/2026/wordml/cei" xmlns:w16cex="http://schemas.microsoft.com/office/word/2018/wordml/cex" xmlns:w16cid="http://schemas.microsoft.com/office/word/2016/wordml/cid" xmlns:w16="http://schemas.microsoft.com/office/word/2018/wordml" xmlns:w16du="http://schemas.microsoft.com/office/word/2023/wordml/word16du" xmlns:w16sdtdh="http://schemas.microsoft.com/office/word/2020/wordml/sdtdatahash" xmlns:w16sdtfl="http://schemas.microsoft.com/office/word/2024/wordml/sdtformatlock" xmlns:w16se="http://schemas.microsoft.com/office/word/2015/wordml/symex" xmlns:wpg="http://schemas.microsoft.com/office/word/2010/wordprocessingGroup" xmlns:wpi="http://schemas.microsoft.com/office/word/2010/wordprocessingInk" xmlns:wne="http://schemas.microsoft.com/office/word/2006/wordml" xmlns:wps="http://schemas.microsoft.com/office/word/2010/wordprocessingShape" mc:Ignorable="w14 w15 w16se w16cid w16 w16cex w16sdtdh w16sdtfl w16du wp14"><w:body>` +
		sections.String() +
		`<w:sectPr w:rsidR="00A948ED" w:rsidSect="00057C3B"><w:headerReference w:type="default" r:id="rId7"/><w:footerReference w:type="default" r:id="rId8"/><w:pgSz w:w="16838" w:h="11906" w:orient="landscape"/><w:pgMar w:top="567" w:right="567" w:bottom="567" w:left="567" w:header="851" w:footer="454" w:gutter="0"/><w:cols w:space="425"/><w:docGrid w:type="lines" w:linePitch="360"/></w:sectPr></w:body></w:document>`
}

func buildAssetInventoryRecordXML(idx int, record models.AssetInventoryRecord) string {
	title := fmt.Sprintf("%02d. %s", idx, firstNonEmptyAssetValue(record.AssetName, record.SystemName, "資訊資產"))
	return paragraphXML(title, "left", true, 24) +
		buildAssetInventoryTableXML(record) +
		paragraphXML("", "left", false, 16)
}

func buildAssetInventoryTableXML(record models.AssetInventoryRecord) string {
	return `<w:tbl><w:tblPr><w:tblW w:w="5000" w:type="pct"/><w:tblBorders><w:top w:val="single" w:sz="8" w:space="0" w:color="auto"/><w:left w:val="single" w:sz="8" w:space="0" w:color="auto"/><w:bottom w:val="single" w:sz="8" w:space="0" w:color="auto"/><w:right w:val="single" w:sz="8" w:space="0" w:color="auto"/><w:insideH w:val="single" w:sz="6" w:space="0" w:color="auto"/><w:insideV w:val="single" w:sz="6" w:space="0" w:color="auto"/></w:tblBorders><w:tblCellMar><w:top w:w="80" w:type="dxa"/><w:left w:w="80" w:type="dxa"/><w:bottom w:w="80" w:type="dxa"/><w:right w:w="80" w:type="dxa"/></w:tblCellMar></w:tblPr><w:tblGrid><w:gridCol w:w="1800"/><w:gridCol w:w="3200"/><w:gridCol w:w="1800"/><w:gridCol w:w="3200"/><w:gridCol w:w="1800"/><w:gridCol w:w="3200"/></w:tblGrid>` +
		tableRowXML(labelCell("資通系統名稱", 1), valueCell(record.SystemName, 5)) +
		tableRowXML(labelCell("資產編號", 1), valueCell(record.AssetCode, 1), labelCell("資產類別", 1), valueCell(record.AssetType, 1), labelCell("資產名稱", 1), valueCell(record.AssetName, 1)) +
		tableRowXML(labelCell("廠牌/廠商", 1), valueCell(record.VendorName, 5)) +
		tableRowXML(labelCell("核心資產", 1), valueCell(record.IsCoreAsset, 1), labelCell("國安疑慮", 1), valueCell(record.HasNationalSecurityConcern, 1), labelCell("數量", 1), valueCell(record.Quantity, 1)) +
		tableRowXML(labelCell("資產說明", 1), valueCell(record.AssetDescription, 5)) +
		tableRowXML(labelCell("管理者(部門)", 1), valueCell(record.ManagerDepartment, 2), labelCell("使用者(部門)", 1), valueCell(record.UserDepartment, 2)) +
		tableRowXML(labelCell("存放位置", 1), valueCell(record.Location, 5)) +
		tableRowXML(labelCell("OS 組態基準", 1), valueCell(record.OsConfigBaseline, 2), labelCell("瀏覽器組態基準", 1), valueCell(record.BrowserConfigBaseline, 2)) +
		tableRowXML(labelCell("網通設備組態基準", 1), valueCell(record.NetworkConfigBaseline, 2), labelCell("應用程式組態基準", 1), valueCell(record.ApplicationConfigBaseline, 2)) +
		tableRowXML(labelCell("其他組態基準", 1), valueCell(record.OtherConfigBaseline, 2), labelCell("組態例外編號", 1), valueCell(record.ConfigExceptionCode, 2)) +
		tableRowXML(labelCell("機密性(A)", 1), valueCell(record.Confidentiality, 1), labelCell("完整性(B)", 1), valueCell(record.Integrity, 1), labelCell("可用性(C)", 1), valueCell(record.Availability, 1)) +
		tableRowXML(labelCell("資產價值 MAX(ABC)", 1), valueCell(record.AssetValue, 1), labelCell("法律遵循性(D)", 1), valueCell(record.LegalCompliance, 1), labelCell("防護等級 MAX(ABCD)", 1), valueCell(record.ProtectionLevel, 1)) +
		tableRowXML(labelCell("MTPD", 1), valueCell(record.Mtpd, 1), labelCell("RTO", 1), valueCell(record.Rto, 1), labelCell("RPO", 1), valueCell(record.Rpo, 1)) +
		tableRowXML(labelCell("狀態", 1), valueCell(record.Status, 2), labelCell("輸入人員", 1), valueCell(record.Creator, 2)) +
		tableRowXML(labelCell("備註", 1), valueCell(record.Remarks, 5)) +
		`</w:tbl>`
}

func firstNonEmptyAssetValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
