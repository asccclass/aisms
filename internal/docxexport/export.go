package docxexport

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"isms-privilege/internal/models"
	"os"
	"path/filepath"
	"strings"
)

const (
	templateHeaderPath   = "word/header1.xml"
	templateDocumentPath = "word/document.xml"
)

type ExportOptions struct {
	TemplatePath string
	FormName     string
	FormCode     string
	Version      string
	RecordCode   string
	Department   string
	InventoryBy  string
	GroupLeader  string
	Accounts     []models.PrivilegedAccount
}

func Generate(opts ExportOptions) ([]byte, error) {
	if opts.TemplatePath == "" {
		return nil, fmt.Errorf("template path is required")
	}
	headerXML := buildHeaderXML(opts)
	documentXML := buildDocumentXML(opts)
	return generateFromTemplate(opts.TemplatePath, headerXML, documentXML)
}

func generateFromTemplate(templatePath, headerXML, documentXML string) ([]byte, error) {
	if templatePath == "" {
		return nil, fmt.Errorf("template path is required")
	}
	src, err := os.Open(templatePath)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	stat, err := src.Stat()
	if err != nil {
		return nil, err
	}

	zr, err := zip.NewReader(src, stat.Size())
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	zw := zip.NewWriter(&out)

	for _, file := range zr.File {
		target, err := zw.CreateHeader(cloneFileHeader(file))
		if err != nil {
			_ = zw.Close()
			return nil, err
		}

		switch filepath.ToSlash(file.Name) {
		case templateHeaderPath:
			if _, err := io.Copy(target, strings.NewReader(headerXML)); err != nil {
				_ = zw.Close()
				return nil, err
			}
			continue
		case templateDocumentPath:
			if _, err := io.Copy(target, strings.NewReader(documentXML)); err != nil {
				_ = zw.Close()
				return nil, err
			}
			continue
		}

		rc, err := file.Open()
		if err != nil {
			_ = zw.Close()
			return nil, err
		}
		if _, err := io.Copy(target, rc); err != nil {
			_ = rc.Close()
			_ = zw.Close()
			return nil, err
		}
		_ = rc.Close()
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func cloneFileHeader(file *zip.File) *zip.FileHeader {
	h := file.FileHeader
	return &h
}

func buildHeaderXML(opts ExportOptions) string {
	formName := defaultString(opts.FormName, "特殊權限帳號盤點清冊")
	formCode := defaultString(opts.FormCode, "ISMS-04-062")
	version := defaultString(opts.Version, "1.4")
	recordCode := strings.TrimSpace(opts.RecordCode)
	department := defaultString(opts.Department, "資安科")

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:hdr xmlns:wpc="http://schemas.microsoft.com/office/word/2010/wordprocessingCanvas" xmlns:cx="http://schemas.microsoft.com/office/drawing/2014/chartex" xmlns:cx1="http://schemas.microsoft.com/office/drawing/2015/9/8/chartex" xmlns:cx2="http://schemas.microsoft.com/office/drawing/2015/10/21/chartex" xmlns:cx3="http://schemas.microsoft.com/office/drawing/2016/5/9/chartex" xmlns:cx4="http://schemas.microsoft.com/office/drawing/2016/5/10/chartex" xmlns:cx5="http://schemas.microsoft.com/office/drawing/2016/5/11/chartex" xmlns:cx6="http://schemas.microsoft.com/office/drawing/2016/5/12/chartex" xmlns:cx7="http://schemas.microsoft.com/office/drawing/2016/5/13/chartex" xmlns:cx8="http://schemas.microsoft.com/office/drawing/2016/5/14/chartex" xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006" xmlns:aink="http://schemas.microsoft.com/office/drawing/2016/ink" xmlns:am3d="http://schemas.microsoft.com/office/drawing/2017/model3d" xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:oel="http://schemas.microsoft.com/office/2019/extlst" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:m="http://schemas.openxmlformats.org/officeDocument/2006/math" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:wp14="http://schemas.microsoft.com/office/word/2010/wordprocessingDrawing" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:w10="urn:schemas-microsoft-com:office:word" xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:w14="http://schemas.microsoft.com/office/word/2010/wordml" xmlns:w15="http://schemas.microsoft.com/office/word/2012/wordml" xmlns:w16cei="http://schemas.microsoft.com/office/word/2026/wordml/cei" xmlns:w16cex="http://schemas.microsoft.com/office/word/2018/wordml/cex" xmlns:w16cid="http://schemas.microsoft.com/office/word/2016/wordml/cid" xmlns:w16="http://schemas.microsoft.com/office/word/2018/wordml" xmlns:w16du="http://schemas.microsoft.com/office/word/2023/wordml/word16du" xmlns:w16sdtdh="http://schemas.microsoft.com/office/word/2020/wordml/sdtdatahash" xmlns:w16sdtfl="http://schemas.microsoft.com/office/word/2024/wordml/sdtformatlock" xmlns:w16se="http://schemas.microsoft.com/office/word/2015/wordml/symex" xmlns:wpg="http://schemas.microsoft.com/office/word/2010/wordprocessingGroup" xmlns:wpi="http://schemas.microsoft.com/office/word/2010/wordprocessingInk" xmlns:wne="http://schemas.microsoft.com/office/word/2006/wordml" xmlns:wps="http://schemas.microsoft.com/office/word/2010/wordprocessingShape" mc:Ignorable="w14 w15 w16se w16cid w16 w16cex w16sdtdh w16sdtfl w16du wp14"><w:tbl><w:tblPr><w:tblStyle w:val="a8"/><w:tblW w:w="4965" w:type="pct"/><w:jc w:val="center"/><w:tblLook w:val="04A0" w:firstRow="1" w:lastRow="0" w:firstColumn="1" w:lastColumn="0" w:noHBand="0" w:noVBand="1"/></w:tblPr><w:tblGrid><w:gridCol w:w="2602"/><w:gridCol w:w="2372"/><w:gridCol w:w="1165"/><w:gridCol w:w="2131"/><w:gridCol w:w="2552"/><w:gridCol w:w="4199"/></w:tblGrid><w:tr w:rsidR="005152D3" w:rsidRPr="005152D3" w14:paraId="243DB228" w14:textId="77777777" w:rsidTr="00057C3B"><w:trPr><w:trHeight w:val="274"/><w:jc w:val="center"/></w:trPr><w:tc><w:tcPr><w:tcW w:w="2602" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr><w:p w14:paraId="02A1CAF6" w14:textId="77777777" w:rsidR="005152D3" w:rsidRPr="005152D3" w:rsidRDefault="005152D3" w:rsidP="005152D3"><w:pPr><w:tabs><w:tab w:val="center" w:pos="4153"/><w:tab w:val="right" w:pos="8306"/></w:tabs><w:adjustRightInd w:val="0"/><w:snapToGrid w:val="0"/><w:spacing w:before="100" w:beforeAutospacing="1" w:after="100" w:afterAutospacing="1"/><w:jc w:val="center"/><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr></w:pPr><w:r w:rsidRPr="005152D3"><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr><w:t>表單或紀錄名稱</w:t></w:r></w:p></w:tc><w:tc><w:tcPr><w:tcW w:w="5668" w:type="dxa"/><w:gridSpan w:val="3"/><w:vAlign w:val="center"/></w:tcPr><w:p w14:paraId="13F3FEAF" w14:textId="0A9E517B" w:rsidR="005152D3" w:rsidRPr="005152D3" w:rsidRDefault="00B338FF" w:rsidP="00DA593F"><w:pPr><w:pStyle w:val="a4"/><w:adjustRightInd w:val="0"/><w:spacing w:before="100" w:beforeAutospacing="1" w:after="100" w:afterAutospacing="1"/><w:jc w:val="center"/><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/></w:rPr></w:pPr><w:r><w:rPr><w:rFonts w:eastAsia="標楷體" w:hint="eastAsia"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/></w:rPr><w:t>%s</w:t></w:r><w:r w:rsidR="005152D3" w:rsidRPr="005152D3"><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/></w:rPr><w:t>%s</w:t></w:r></w:p></w:tc><w:tc><w:tcPr><w:tcW w:w="2552" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr><w:p w14:paraId="127E45B6" w14:textId="77777777" w:rsidR="005152D3" w:rsidRPr="005152D3" w:rsidRDefault="005152D3" w:rsidP="005152D3"><w:pPr><w:tabs><w:tab w:val="center" w:pos="4153"/><w:tab w:val="right" w:pos="8306"/></w:tabs><w:adjustRightInd w:val="0"/><w:snapToGrid w:val="0"/><w:spacing w:before="100" w:beforeAutospacing="1" w:after="100" w:afterAutospacing="1"/><w:jc w:val="center"/><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr></w:pPr><w:r w:rsidRPr="005152D3"><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr><w:t>表單或紀錄編號</w:t></w:r></w:p></w:tc><w:tc><w:tcPr><w:tcW w:w="4199" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr><w:p w14:paraId="11452233" w14:textId="7782B82C" w:rsidR="005152D3" w:rsidRPr="005152D3" w:rsidRDefault="00306FB1" w:rsidP="005152D3"><w:pPr><w:tabs><w:tab w:val="center" w:pos="4153"/><w:tab w:val="right" w:pos="8306"/></w:tabs><w:adjustRightInd w:val="0"/><w:snapToGrid w:val="0"/><w:spacing w:before="100" w:beforeAutospacing="1" w:after="100" w:afterAutospacing="1"/><w:jc w:val="center"/><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr></w:pPr><w:r w:rsidRPr="00306FB1"><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr><w:t>%s</w:t></w:r></w:p></w:tc></w:tr><w:tr w:rsidR="005152D3" w:rsidRPr="005152D3" w14:paraId="233127A7" w14:textId="77777777" w:rsidTr="00057C3B"><w:trPr><w:trHeight w:val="286"/><w:jc w:val="center"/></w:trPr><w:tc><w:tcPr><w:tcW w:w="2602" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr><w:p w14:paraId="6DE861C6" w14:textId="77777777" w:rsidR="005152D3" w:rsidRPr="005152D3" w:rsidRDefault="005152D3" w:rsidP="005152D3"><w:pPr><w:pStyle w:val="a4"/><w:adjustRightInd w:val="0"/><w:spacing w:before="100" w:beforeAutospacing="1" w:after="100" w:afterAutospacing="1"/><w:jc w:val="center"/><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/></w:rPr></w:pPr><w:r w:rsidRPr="005152D3"><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/></w:rPr><w:t>機密等級</w:t></w:r></w:p></w:tc><w:tc><w:tcPr><w:tcW w:w="2372" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr><w:p w14:paraId="06FBE072" w14:textId="77777777" w:rsidR="005152D3" w:rsidRPr="005152D3" w:rsidRDefault="005152D3" w:rsidP="005152D3"><w:pPr><w:pStyle w:val="a4"/><w:adjustRightInd w:val="0"/><w:spacing w:before="100" w:beforeAutospacing="1" w:after="100" w:afterAutospacing="1"/><w:jc w:val="center"/><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/></w:rPr></w:pPr><w:r w:rsidRPr="005152D3"><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/></w:rPr><w:t>限</w:t></w:r></w:p></w:tc><w:tc><w:tcPr><w:tcW w:w="1165" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr><w:p w14:paraId="07DE1E1B" w14:textId="77777777" w:rsidR="005152D3" w:rsidRPr="005152D3" w:rsidRDefault="005152D3" w:rsidP="005152D3"><w:pPr><w:pStyle w:val="a4"/><w:adjustRightInd w:val="0"/><w:spacing w:before="100" w:beforeAutospacing="1" w:after="100" w:afterAutospacing="1"/><w:jc w:val="center"/><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/></w:rPr></w:pPr><w:r w:rsidRPr="005152D3"><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/></w:rPr><w:t>版本</w:t></w:r></w:p></w:tc><w:tc><w:tcPr><w:tcW w:w="2131" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr><w:p w14:paraId="6F35971F" w14:textId="0A422CFF" w:rsidR="005152D3" w:rsidRPr="005152D3" w:rsidRDefault="005152D3" w:rsidP="00DA593F"><w:pPr><w:pStyle w:val="a4"/><w:adjustRightInd w:val="0"/><w:spacing w:before="100" w:beforeAutospacing="1" w:after="100" w:afterAutospacing="1"/><w:jc w:val="center"/><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/></w:rPr></w:pPr><w:r w:rsidRPr="005152D3"><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/></w:rPr><w:t>%s</w:t></w:r></w:p></w:tc><w:tc><w:tcPr><w:tcW w:w="2552" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr><w:p w14:paraId="75F36BA2" w14:textId="77777777" w:rsidR="005152D3" w:rsidRPr="005152D3" w:rsidRDefault="005152D3" w:rsidP="005152D3"><w:pPr><w:tabs><w:tab w:val="center" w:pos="4153"/><w:tab w:val="right" w:pos="8306"/></w:tabs><w:adjustRightInd w:val="0"/><w:snapToGrid w:val="0"/><w:spacing w:before="100" w:beforeAutospacing="1" w:after="100" w:afterAutospacing="1"/><w:jc w:val="center"/><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr></w:pPr><w:r w:rsidRPr="005152D3"><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr><w:t>權責單位</w:t></w:r></w:p></w:tc><w:tc><w:tcPr><w:tcW w:w="4199" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr><w:p w14:paraId="16750AEF" w14:textId="007962F9" w:rsidR="005152D3" w:rsidRPr="005152D3" w:rsidRDefault="00FF74AF" w:rsidP="005152D3"><w:pPr><w:tabs><w:tab w:val="center" w:pos="4153"/><w:tab w:val="right" w:pos="8306"/></w:tabs><w:adjustRightInd w:val="0"/><w:snapToGrid w:val="0"/><w:spacing w:before="100" w:beforeAutospacing="1" w:after="100" w:afterAutospacing="1"/><w:jc w:val="center"/><w:rPr><w:rFonts w:eastAsia="標楷體"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr></w:pPr><w:r><w:rPr><w:rFonts w:eastAsia="標楷體" w:hint="eastAsia"/><w:color w:val="808080" w:themeColor="background1" w:themeShade="80"/><w:sz w:val="20"/><w:szCs w:val="20"/></w:rPr><w:t>%s</w:t></w:r></w:p></w:tc></w:tr></w:tbl><w:p w14:paraId="2FCAE751" w14:textId="77777777" w:rsidR="005152D3" w:rsidRDefault="005152D3"><w:pPr><w:pStyle w:val="a4"/></w:pPr></w:p></w:hdr>`,
		xmlEscape(formCode),
		xmlEscape(formName),
		xmlEscape(recordCode),
		xmlEscape(version),
		xmlEscape(department),
	)
}

func buildDocumentXML(opts ExportOptions) string {
	var rows strings.Builder
	for _, a := range opts.Accounts {
		rows.WriteString(buildAccountRowXML(a))
	}
	if len(opts.Accounts) == 0 {
		rows.WriteString(buildEmptyRowXML())
	}

	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<w:document xmlns:wpc="http://schemas.microsoft.com/office/word/2010/wordprocessingCanvas" xmlns:cx="http://schemas.microsoft.com/office/drawing/2014/chartex" xmlns:cx1="http://schemas.microsoft.com/office/drawing/2015/9/8/chartex" xmlns:cx2="http://schemas.microsoft.com/office/drawing/2015/10/21/chartex" xmlns:cx3="http://schemas.microsoft.com/office/drawing/2016/5/9/chartex" xmlns:cx4="http://schemas.microsoft.com/office/drawing/2016/5/10/chartex" xmlns:cx5="http://schemas.microsoft.com/office/drawing/2016/5/11/chartex" xmlns:cx6="http://schemas.microsoft.com/office/drawing/2016/5/12/chartex" xmlns:cx7="http://schemas.microsoft.com/office/drawing/2016/5/13/chartex" xmlns:cx8="http://schemas.microsoft.com/office/drawing/2016/5/14/chartex" xmlns:mc="http://schemas.openxmlformats.org/markup-compatibility/2006" xmlns:aink="http://schemas.microsoft.com/office/drawing/2016/ink" xmlns:am3d="http://schemas.microsoft.com/office/drawing/2017/model3d" xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:oel="http://schemas.microsoft.com/office/2019/extlst" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:m="http://schemas.openxmlformats.org/officeDocument/2006/math" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:wp14="http://schemas.microsoft.com/office/word/2010/wordprocessingDrawing" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:w10="urn:schemas-microsoft-com:office:word" xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:w14="http://schemas.microsoft.com/office/word/2010/wordml" xmlns:w15="http://schemas.microsoft.com/office/word/2012/wordml" xmlns:w16cei="http://schemas.microsoft.com/office/word/2026/wordml/cei" xmlns:w16cex="http://schemas.microsoft.com/office/word/2018/wordml/cex" xmlns:w16cid="http://schemas.microsoft.com/office/word/2016/wordml/cid" xmlns:w16="http://schemas.microsoft.com/office/word/2018/wordml" xmlns:w16du="http://schemas.microsoft.com/office/word/2023/wordml/word16du" xmlns:w16sdtdh="http://schemas.microsoft.com/office/word/2020/wordml/sdtdatahash" xmlns:w16sdtfl="http://schemas.microsoft.com/office/word/2024/wordml/sdtformatlock" xmlns:w16se="http://schemas.microsoft.com/office/word/2015/wordml/symex" xmlns:wpg="http://schemas.microsoft.com/office/word/2010/wordprocessingGroup" xmlns:wpi="http://schemas.microsoft.com/office/word/2010/wordprocessingInk" xmlns:wne="http://schemas.microsoft.com/office/word/2006/wordml" xmlns:wps="http://schemas.microsoft.com/office/word/2010/wordprocessingShape" mc:Ignorable="w14 w15 w16se w16cid w16 w16cex w16sdtdh w16sdtfl w16du wp14"><w:body>` +
		buildMainTableHeaderXML() +
		rows.String() +
		`</w:tbl><w:p w14:paraId="717F5FE6" w14:textId="7C1F7432" w:rsidR="00190128" w:rsidRPr="00D727C4" w:rsidRDefault="00D727C4" w:rsidP="00C07ACE"><w:pPr><w:widowControl/><w:rPr><w:rFonts w:ascii="標楷體" w:eastAsia="標楷體" w:hAnsi="標楷體"/></w:rPr></w:pPr><w:r><w:rPr><w:rFonts w:ascii="微軟正黑體" w:eastAsia="微軟正黑體" w:hAnsi="微軟正黑體" w:hint="eastAsia"/></w:rPr><w:t>★</w:t></w:r><w:r><w:rPr><w:rFonts w:ascii="標楷體" w:eastAsia="標楷體" w:hAnsi="標楷體" w:hint="eastAsia"/></w:rPr><w:t>若為建置初始所預設之帳號，</w:t></w:r><w:r><w:rPr><w:rFonts w:ascii="標楷體" w:eastAsia="標楷體" w:hAnsi="標楷體" w:hint="eastAsia"/><w:u w:val="single"/></w:rPr><w:t>帳號種類</w:t></w:r><w:r><w:rPr><w:rFonts w:ascii="標楷體" w:eastAsia="標楷體" w:hAnsi="標楷體" w:hint="eastAsia"/></w:rPr><w:t>請填「</w:t></w:r><w:r><w:rPr><w:rFonts w:ascii="標楷體" w:eastAsia="標楷體" w:hAnsi="標楷體" w:hint="eastAsia"/><w:b/><w:bCs/></w:rPr><w:t>預設</w:t></w:r><w:r><w:rPr><w:rFonts w:ascii="標楷體" w:eastAsia="標楷體" w:hAnsi="標楷體" w:hint="eastAsia"/></w:rPr><w:t>」</w:t></w:r></w:p>` +
		buildSignatureTableXML(opts.InventoryBy, opts.GroupLeader) +
		`<w:p w14:paraId="7231F696" w14:textId="77777777" w:rsidR="00A948ED" w:rsidRDefault="00A948ED" w:rsidP="00057C3B"><w:pPr><w:spacing w:line="20" w:lineRule="exact"/></w:pPr></w:p><w:sectPr w:rsidR="00A948ED" w:rsidSect="00057C3B"><w:headerReference w:type="default" r:id="rId7"/><w:footerReference w:type="default" r:id="rId8"/><w:pgSz w:w="16838" w:h="11906" w:orient="landscape"/><w:pgMar w:top="567" w:right="1134" w:bottom="567" w:left="567" w:header="851" w:footer="454" w:gutter="0"/><w:cols w:space="425"/><w:docGrid w:type="lines" w:linePitch="360"/></w:sectPr></w:body></w:document>`
}

func buildMainTableHeaderXML() string {
	return `<w:tbl><w:tblPr><w:tblW w:w="15026" w:type="dxa"/><w:tblInd w:w="-15" w:type="dxa"/><w:tblBorders><w:top w:val="thinThickMediumGap" w:sz="24" w:space="0" w:color="auto"/><w:left w:val="thinThickMediumGap" w:sz="24" w:space="0" w:color="auto"/><w:bottom w:val="thickThinMediumGap" w:sz="24" w:space="0" w:color="auto"/><w:right w:val="thickThinMediumGap" w:sz="24" w:space="0" w:color="auto"/><w:insideH w:val="single" w:sz="4" w:space="0" w:color="auto"/><w:insideV w:val="single" w:sz="4" w:space="0" w:color="auto"/></w:tblBorders><w:tblCellMar><w:left w:w="28" w:type="dxa"/><w:right w:w="28" w:type="dxa"/></w:tblCellMar><w:tblLook w:val="0000" w:firstRow="0" w:lastRow="0" w:firstColumn="0" w:lastColumn="0" w:noHBand="0" w:noVBand="0"/></w:tblPr><w:tblGrid><w:gridCol w:w="1958"/><w:gridCol w:w="2105"/><w:gridCol w:w="1545"/><w:gridCol w:w="1541"/><w:gridCol w:w="1393"/><w:gridCol w:w="1486"/><w:gridCol w:w="1622"/><w:gridCol w:w="3376"/></w:tblGrid><w:tr w:rsidR="001367CA" w14:paraId="77D76BD0" w14:textId="77777777" w:rsidTr="00787E93"><w:trPr><w:cantSplit/><w:trHeight w:val="750"/></w:trPr>` +
		headerCell("1958", "系統名稱") +
		headerCell("2105", "IP位址") +
		headerCell("1545", "盤點日期") +
		headerCell("1541", "帳號名稱") +
		headerCell("1393", "帳號種類") +
		headerCell("1486", "單位/姓名") +
		`<w:tc><w:tcPr><w:tcW w:w="1622" w:type="dxa"/><w:shd w:val="clear" w:color="auto" w:fill="D9D9D9"/></w:tcPr><w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:rFonts w:eastAsia="標楷體" w:hint="eastAsia"/><w:b/><w:bCs/></w:rPr><w:t>通行碼(含PassPhrase)</w:t></w:r></w:p><w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:rFonts w:eastAsia="標楷體" w:hint="eastAsia"/><w:b/><w:bCs/></w:rPr><w:t>是否定期變更</w:t></w:r></w:p></w:tc>` +
		headerCell("3376", "備註") +
		`</w:tr>`
}

func headerCell(width, text string) string {
	return `<w:tc><w:tcPr><w:tcW w:w="` + width + `" w:type="dxa"/><w:shd w:val="clear" w:color="auto" w:fill="D9D9D9"/><w:vAlign w:val="center"/></w:tcPr><w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:rFonts w:eastAsia="標楷體" w:hint="eastAsia"/><w:b/><w:bCs/></w:rPr><w:t>` + xmlEscape(text) + `</w:t></w:r></w:p></w:tc>`
}

func buildAccountRowXML(a models.PrivilegedAccount) string {
	unitOwner := strings.TrimSpace(strings.Trim(strings.Join([]string{strings.TrimSpace(a.Department), strings.TrimSpace(a.OwnerName)}, "/"), "/"))
	return `<w:tr><w:trPr><w:cantSplit/><w:trHeight w:val="750"/></w:trPr>` +
		dataCell("1958", a.SystemName, "center") +
		dataCell("2105", a.IPAddress, "center") +
		dataCell("1545", a.InventoryDate, "center") +
		dataCell("1541", a.AccountName, "center") +
		dataCell("1393", a.AccountType, "center") +
		dataCell("1486", unitOwner, "center") +
		dataCell("1622", a.PassphraseRotate, "right") +
		dataCell("3376", a.Remarks, "left") +
		`</w:tr>`
}

func buildEmptyRowXML() string {
	return `<w:tr><w:trPr><w:cantSplit/><w:trHeight w:val="750"/></w:trPr>` +
		dataCell("1958", "", "center") +
		dataCell("2105", "", "center") +
		dataCell("1545", "", "center") +
		dataCell("1541", "", "center") +
		dataCell("1393", "", "center") +
		dataCell("1486", "", "center") +
		dataCell("1622", "", "right") +
		dataCell("3376", "", "left") +
		`</w:tr>`
}

func dataCell(width, text, align string) string {
	if align == "" {
		align = "left"
	}
	return `<w:tc><w:tcPr><w:tcW w:w="` + width + `" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr><w:p><w:pPr><w:spacing w:line="480" w:lineRule="auto"/><w:jc w:val="` + align + `"/></w:pPr>` + textRunXML(text) + `</w:p></w:tc>`
}

func textRunXML(text string) string {
	if text == "" {
		return ""
	}
	return `<w:r><w:rPr><w:rFonts w:ascii="Times New Roman" w:hint="eastAsia"/><w:sz w:val="22"/></w:rPr><w:t>` + xmlEscape(text) + `</w:t></w:r>`
}

func buildSignatureTableXML(inventoryBy, groupLeader string) string {
	return `<w:tbl><w:tblPr><w:tblW w:w="15168" w:type="dxa"/><w:tblInd w:w="-15" w:type="dxa"/><w:tblBorders><w:top w:val="single" w:sz="4" w:space="0" w:color="auto"/><w:left w:val="single" w:sz="4" w:space="0" w:color="auto"/><w:bottom w:val="single" w:sz="4" w:space="0" w:color="auto"/><w:right w:val="single" w:sz="4" w:space="0" w:color="auto"/><w:insideH w:val="single" w:sz="4" w:space="0" w:color="auto"/><w:insideV w:val="single" w:sz="4" w:space="0" w:color="auto"/></w:tblBorders><w:tblLook w:val="01E0" w:firstRow="1" w:lastRow="1" w:firstColumn="1" w:lastColumn="1" w:noHBand="0" w:noVBand="0"/></w:tblPr><w:tblGrid><w:gridCol w:w="7513"/><w:gridCol w:w="7655"/></w:tblGrid><w:tr><w:trPr><w:trHeight w:val="510"/></w:trPr>` +
		headerCell("7513", "盤點者") +
		headerCell("7655", "組長") +
		`</w:tr><w:tr><w:trPr><w:trHeight w:val="575"/></w:trPr>` +
		dataCell("7513", inventoryBy, "center") +
		dataCell("7655", groupLeader, "center") +
		`</w:tr></w:tbl>`
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func xmlEscape(value string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(value))
	return b.String()
}
