package db

import (
	"database/sql"
	"isms-privilege/internal/models"
	"strings"
)

type baselineTemplateControl struct {
	ItemNo             int
	DomainName         string
	ControlCategory    string
	RequirementLevel   string
	ControlDescription string
	MeasureNotes       string
	Remarks            string
}

var protectionBaselineTemplate = []baselineTemplateControl{
	{1, "存取控制", "帳號管理", "普", "建立帳號管理機制，包含帳號之申請、建立、修改、啟用、停用及刪除之程序。", "建立帳號管理機制，包含帳號之申請、建立、修改、啟用、停用及刪除之程序。", ""},
	{2, "存取控制", "帳號管理", "普", "已逾期之臨時或緊急帳號應刪除或禁用。", "", "原中級移列普級規範。"},
	{3, "存取控制", "帳號管理", "普", "資通系統閒置帳號應禁用。", "", "原中級移列普級規範。"},
	{4, "存取控制", "帳號管理", "普", "定期審核資通系統帳號之申請、建立、修改、啟用、停用及刪除。", "", "原中級移列普級規範。"},
	{5, "存取控制", "帳號管理", "中", "機關應定義各系統之閒置時間或可使用期限與資通系統之使用情況及條件。", "", "原高級移列中級規範。"},
	{6, "存取控制", "帳號管理", "中", "逾越機關所許可之閒置時間或可使用期限時，系統應自動將使用者登出。", "", "原高級移列中級規範。"},
	{7, "存取控制", "帳號管理", "高", "應依機關規定之情況及條件，使用資通系統。", "", ""},
	{8, "存取控制", "帳號管理", "高", "監控資通系統帳號，如發現帳號違常使用時回報管理者。", "", ""},
	{9, "存取控制", "最小權限", "普", "採最小權限原則，僅允許使用者（或代表使用者行為之程序）依機關任務及業務功能，完成指派任務所需之授權存取。", "", "原中級移列普級規範。"},
	{10, "存取控制", "遠端存取", "普", "對於每一種允許之遠端存取類型，均應先取得授權，建立使用限制、組態需求、連線需求及文件化。", "", ""},
	{11, "存取控制", "遠端存取", "普", "使用者之權限檢查作業應於伺服器端完成。", "", ""},
	{12, "存取控制", "遠端存取", "普", "應監控遠端存取機關內部網段或資通系統後臺之連線。", "", ""},
	{13, "存取控制", "遠端存取", "普", "應採用加密機制。", "", ""},
	{14, "存取控制", "遠端存取", "普", "遠端存取之來源應為機關已預先定義及管理之存取控制點。", "", "原中級移列普級規範。"},
	{15, "事件日誌與可歸責性", "記錄事件", "普", "訂定日誌之記錄時間週期及留存政策，並保留日誌至少六個月。", "", ""},
	{16, "事件日誌與可歸責性", "記錄事件", "普", "確保資通系統有記錄特定事件之功能，並決定應記錄之特定資通系統事件。", "", ""},
	{17, "事件日誌與可歸責性", "記錄事件", "普", "應記錄資通系統管理者帳號所執行之各項功能。", "", ""},
	{18, "事件日誌與可歸責性", "記錄事件", "中", "應定期審查機關所保留資通系統產生之日誌。", "", ""},
	{19, "事件日誌與可歸責性", "日誌紀錄內容", "普", "資通系統產生之日誌應包含事件類型、發生時間、發生位置及任何與事件相關之使用者身分識別等資訊。", "", "調整構面文字"},
	{20, "事件日誌與可歸責性", "日誌儲存容量", "普", "依據日誌儲存需求，配置所需之儲存容量。", "", "調整構面文字"},
	{21, "事件日誌與可歸責性", "日誌處理失效之回應", "普", "資通系統於日誌處理失效時，應採取適當之行動。", "", "調整構面文字"},
	{22, "事件日誌與可歸責性", "日誌處理失效之回應", "高", "機關規定需要即時通報之日誌處理失效事件發生時，資通系統應於機關規定之時效內，對特定人員提出警告。", "", "調整構面文字"},
	{23, "事件日誌與可歸責性", "時戳及校時", "普", "資通系統應使用系統內部時鐘產生日誌所需時戳，並可以對應到世界協調時間(UTC)或格林威治標準時間(GMT)。", "", "調整構面文字"},
	{24, "事件日誌與可歸責性", "時戳及校時", "普", "系統內部時鐘應定期與基準時間源進行同步。", "", "原中級移列普級規範。調整構面文字"},
	{25, "事件日誌與可歸責性", "日誌資訊之保護", "普", "對日誌之存取管理，僅限於有權限之使用者。", "", "調整構面文字"},
	{26, "事件日誌與可歸責性", "日誌資訊之保護", "中", "應運用雜湊或其他適當方式之完整性確保機制。", "", "調整構面文字"},
	{27, "事件日誌與可歸責性", "日誌資訊之保護", "高", "定期備份日誌至原系統外之其他實體系統。", "", "調整構面文字"},
	{28, "營運持續計畫", "資料備份", "普", "訂定資料可容忍損失之時間要求。", "", "營運持續計畫「系統備份」修正為「資料備份」"},
	{29, "營運持續計畫", "資料備份", "普", "執行資料備份。", "", "營運持續計畫「系統備份」修正為「資料備份」"},
	{30, "營運持續計畫", "資料備份", "中", "應定期測試備份資料，以驗證備份媒體之可靠性及資訊之完整性。", "", "營運持續計畫「系統備份」修正為「資料備份」"},
	{31, "營運持續計畫", "資料備份", "高", "應將備份還原，作為營運持續計畫演練之一部分。", "", "營運持續計畫「系統備份」修正為「資料備份」"},
	{32, "營運持續計畫", "資料備份", "高", "應建立資料異地備份機制。", "", "營運持續計畫「系統備份」修正為「資料備份」"},
	{33, "營運持續計畫", "系統備援", "普", "訂定資通系統從中斷後至重新恢復服務之可容忍時間要求。", "", ""},
	{34, "營運持續計畫", "系統備援", "中", "應定期測試原服務中斷時，於最大可容忍中斷時間內，由備援設備或其他方式取代並提供服務。", "", ""},
	{35, "營運持續計畫", "系統備援", "高", "應將備援啟動作為營運持續計畫演練之一部分。", "", ""},
	{36, "識別與鑑別", "使用者之識別與鑑別", "普", "資通系統應識別及鑑別使用者，並禁止使用者使用共用帳號。", "", ""},
	{37, "識別與鑑別", "使用者之識別與鑑別", "高", "對資通系統之存取採取多因子鑑別技術。", "", ""},
	{38, "識別與鑑別", "身分驗證管理", "普", "使用預設密碼初次登入系統時，應於登入後要求立即變更。", "", ""},
	{39, "識別與鑑別", "身分驗證管理", "普", "身分驗證相關資訊不以明文傳輸。", "", ""},
	{40, "識別與鑑別", "身分驗證管理", "普", "具備帳戶鎖定機制，帳號登入進行身分驗證失敗達五次後，至少十五分鐘內不允許該帳號繼續嘗試登入或使用機關自建之失敗驗證機制。", "", ""},
	{41, "識別與鑑別", "身分驗證管理", "普", "使用密碼進行驗證時，應強制最低密碼複雜度；依機關密碼效期規定變更密碼。", "", ""},
	{42, "識別與鑑別", "身分驗證管理", "普", "密碼變更時，至少不可以與前三次使用過之密碼相同。", "", ""},
	{43, "識別與鑑別", "身分驗證管理", "普", "第四點及第五點所定措施，對外部使用者，機關得自行規範辦理。", "", ""},
	{44, "識別與鑑別", "身分驗證管理", "中", "身分驗證機制應防範自動化程式之登入或密碼更換嘗試。", "", ""},
	{45, "識別與鑑別", "身分驗證管理", "中", "密碼重設機制對使用者重新身分確認後，發送一次性及具有時效性符記。", "", ""},
	{46, "識別與鑑別", "鑑別資訊保護", "普", "資通系統應遮蔽鑑別過程中之資訊。", "", ""},
	{47, "識別與鑑別", "鑑別資訊保護", "中", "資通系統如以密碼進行鑑別時，該密碼應經雜湊或其他適當方式處理後儲存。", "", ""},
	{48, "系統與服務獲得", "系統發展生命週期需求階段", "普", "針對系統安全需求（含機密性、可用性、完整性）進行確認。", "", ""},
	{49, "系統與服務獲得", "系統發展生命週期設計階段", "中", "根據系統功能與要求，識別可能影響系統之威脅，進行風險分析及評估。", "", ""},
	{50, "系統與服務獲得", "系統發展生命週期設計階段", "中", "將風險評估結果回饋需求階段之檢核項目，並提出安全需求修正。", "", ""},
	{51, "系統與服務獲得", "系統發展生命週期開發階段", "普", "應針對安全需求實作必要控制措施。", "", ""},
	{52, "系統與服務獲得", "系統發展生命週期開發階段", "普", "應注意避免軟體常見漏洞及實作必要控制措施。", "", ""},
	{53, "系統與服務獲得", "系統發展生命週期開發階段", "普", "發生錯誤時，使用者頁面僅顯示簡短錯誤訊息及代碼，不包含詳細之錯誤訊息。", "", ""},
	{54, "系統與服務獲得", "系統發展生命週期開發階段", "高", "執行「源碼掃描」安全檢測。", "", ""},
	{55, "系統與服務獲得", "系統發展生命週期開發階段", "高", "系統應具備發生嚴重錯誤時之通知機制。", "", ""},
	{56, "系統與服務獲得", "系統發展生命週期測試階段", "普", "執行「弱點掃描」安全檢測。", "", ""},
	{57, "系統與服務獲得", "系統發展生命週期測試階段", "高", "執行「滲透測試」安全檢測。", "", ""},
	{58, "系統與服務獲得", "系統發展生命週期部署與維運階段", "普", "於部署環境中應針對相關資通安全威脅，進行更新與修補。", "", ""},
	{59, "系統與服務獲得", "系統發展生命週期部署與維運階段", "普", "識別並關閉不必要服務及埠口。", "", ""},
	{60, "系統與服務獲得", "系統發展生命週期部署與維運階段", "普", "資通系統不使用預設密碼。", "", ""},
	{61, "系統與服務獲得", "系統發展生命週期部署與維運階段", "普", "執行系統源碼備份。", "", ""},
	{62, "系統與服務獲得", "系統發展生命週期部署與維運階段", "中", "於系統發展生命週期之維運階段，應執行版本控制與變更管理。", "", ""},
	{63, "系統與服務獲得", "系統發展生命週期委外階段", "普", "資通系統開發如委外辦理，應將系統發展生命週期各階段依等級將安全需求（含機密性、可用性、完整性）納入委外契約。", "", ""},
	{64, "系統與服務獲得", "獲得程序", "普", "識別資通系統使用之第三方軟體、服務、函式庫或其他元件。", "", ""},
	{65, "系統與服務獲得", "獲得程序", "中", "開發、測試及正式作業環境應為區隔。", "", ""},
	{66, "系統與服務獲得", "系統文件", "普", "應儲存與管理系統發展生命週期之相關文件。", "", ""},
	{67, "系統與通訊保護", "傳輸之機密性與完整性", "高", "資通系統應採用加密機制，以防止未授權之資訊揭露或偵測資訊之變更。", "", ""},
	{68, "系統與通訊保護", "傳輸之機密性與完整性", "高", "使用公開、國際機構驗證且未遭破解之演算法。", "", ""},
	{69, "系統與通訊保護", "傳輸之機密性與完整性", "高", "加密金鑰或憑證應定期更換。", "", ""},
	{70, "系統與通訊保護", "傳輸之機密性與完整性", "高", "伺服器端之金鑰保管應訂定管理規範及實施應有之安全防護措施。", "", ""},
	{71, "系統與通訊保護", "資料儲存之安全", "高", "資通系統重要組態設定檔案及其他具保護需求之資訊應加密或以其他適當方式儲存。", "", ""},
	{72, "系統與資訊完整性", "漏洞修復", "普", "系統之漏洞修復應測試有效性及潛在影響，並定期更新。", "", ""},
	{73, "系統與資訊完整性", "漏洞修復", "中", "定期確認資通系統相關漏洞修復之狀態。", "", ""},
	{74, "系統與資訊完整性", "資通系統監控", "普", "發現資通系統有被入侵跡象時，應通報機關特定人員。", "", ""},
	{75, "系統與資訊完整性", "資通系統監控", "中", "監控資通系統，以偵測攻擊與未授權之連線，並識別資通系統之未授權使用。", "", ""},
	{76, "系統與資訊完整性", "資通系統監控", "高", "資通系統應採用自動化工具監控進出之通信流量，並於發現不尋常或未授權之活動時，針對該事件進行分析。", "", ""},
	{77, "系統與資訊完整性", "軟體及資訊完整性", "普", "使用者輸入資料合法性檢查應置放於應用系統伺服器端。", "", "原中級移列普級規範。"},
	{78, "系統與資訊完整性", "軟體及資訊完整性", "中", "使用完整性驗證工具，以偵測未授權變更特定軟體及資訊。", "", ""},
	{79, "系統與資訊完整性", "軟體及資訊完整性", "中", "發現違反完整性時，資通系統應實施機關指定之安全保護措施。", "", ""},
	{80, "系統與資訊完整性", "軟體及資訊完整性", "高", "應定期執行軟體與資訊完整性檢查。", "", ""},
}

func ProtectionBaselineTemplateControls() []models.ProtectionBaselineControl {
	controls := make([]models.ProtectionBaselineControl, 0, len(protectionBaselineTemplate))
	for _, item := range protectionBaselineTemplate {
		controls = append(controls, models.ProtectionBaselineControl{
			ItemNo:             item.ItemNo,
			DomainName:         item.DomainName,
			ControlCategory:    item.ControlCategory,
			RequirementLevel:   item.RequirementLevel,
			ControlDescription: item.ControlDescription,
			MeasureNotes:       item.MeasureNotes,
			Remarks:            item.Remarks,
		})
	}
	return controls
}

func (d *DB) ListProtectionBaselineRecords() ([]models.ProtectionBaselineRecord, error) {
	return d.ListProtectionBaselineRecordsByCreator("")
}

func (d *DB) ListProtectionBaselineRecordsByCreator(creator string) ([]models.ProtectionBaselineRecord, error) {
	query := `SELECT r.id,r.system_name,r.security_level,r.filled_by,r.form_date,r.reviewer,r.review_date,r.status,r.creator,r.remarks,r.created_at,r.updated_at,
	COUNT(c.id),
	SUM(CASE WHEN c.applies='Y' THEN 1 ELSE 0 END),
	SUM(CASE WHEN c.compliance='符合' THEN 1 ELSE 0 END)
	FROM protection_baseline_records r
	LEFT JOIN protection_baseline_controls c ON c.record_id=r.id`
	args := []interface{}{}
	if strings.TrimSpace(creator) != "" {
		query += ` WHERE r.creator=?`
		args = append(args, creator)
	}
	query += `
	GROUP BY r.id
	ORDER BY r.id DESC`
	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.ProtectionBaselineRecord{}
	for rows.Next() {
		var item models.ProtectionBaselineRecord
		var total, applied, compliant sql.NullInt64
		if err := rows.Scan(&item.ID, &item.SystemName, &item.SecurityLevel, &item.FilledBy, &item.FormDate, &item.Reviewer, &item.ReviewDate, &item.Status, &item.Creator, &item.Remarks, &item.CreatedAt, &item.UpdatedAt, &total, &applied, &compliant); err != nil {
			return nil, err
		}
		item.TotalControls = int(total.Int64)
		item.AppliedCount = int(applied.Int64)
		item.CompliantCount = int(compliant.Int64)
		list = append(list, item)
	}
	return list, nil
}

func (d *DB) GetProtectionBaselineRecord(id int) (*models.ProtectionBaselineRecord, error) {
	return d.GetProtectionBaselineRecordByCreator(id, "")
}

func (d *DB) GetProtectionBaselineRecordByCreator(id int, creator string) (*models.ProtectionBaselineRecord, error) {
	query := `SELECT id,system_name,security_level,filled_by,form_date,reviewer,review_date,status,creator,remarks,created_at,updated_at FROM protection_baseline_records WHERE id=?`
	args := []interface{}{id}
	if strings.TrimSpace(creator) != "" {
		query += ` AND creator=?`
		args = append(args, creator)
	}
	row := d.conn.QueryRow(query, args...)
	var item models.ProtectionBaselineRecord
	if err := row.Scan(&item.ID, &item.SystemName, &item.SecurityLevel, &item.FilledBy, &item.FormDate, &item.Reviewer, &item.ReviewDate, &item.Status, &item.Creator, &item.Remarks, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	controls, err := d.listProtectionBaselineControls(id)
	if err != nil {
		return nil, err
	}
	item.Controls = controls
	item.TotalControls = len(controls)
	for _, c := range controls {
		if strings.TrimSpace(c.Applies) == "Y" {
			item.AppliedCount++
		}
		if strings.TrimSpace(c.Compliance) == "符合" {
			item.CompliantCount++
		}
	}
	return &item, nil
}

func (d *DB) listProtectionBaselineControls(recordID int) ([]models.ProtectionBaselineControl, error) {
	rows, err := d.conn.Query(`SELECT id,record_id,item_no,domain_name,control_category,requirement_level,control_description,measure_notes,applies,implementation_notes,compliance,finding,remarks FROM protection_baseline_controls WHERE record_id=? ORDER BY item_no ASC,id ASC`, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.ProtectionBaselineControl{}
	for rows.Next() {
		var item models.ProtectionBaselineControl
		if err := rows.Scan(&item.ID, &item.RecordID, &item.ItemNo, &item.DomainName, &item.ControlCategory, &item.RequirementLevel, &item.ControlDescription, &item.MeasureNotes, &item.Applies, &item.ImplementationNotes, &item.Compliance, &item.Finding, &item.Remarks); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (d *DB) CreateProtectionBaselineRecord(r *models.ProtectionBaselineRecord) (int64, error) {
	tx, err := d.conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`INSERT INTO protection_baseline_records (system_name,security_level,filled_by,form_date,reviewer,review_date,status,creator,remarks) VALUES (?,?,?,?,?,?,?,?,?)`,
		r.SystemName, r.SecurityLevel, r.FilledBy, r.FormDate, r.Reviewer, r.ReviewDate, r.Status, r.Creator, r.Remarks)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	controls := r.Controls
	if len(controls) == 0 {
		controls = ProtectionBaselineTemplateControls()
	}
	for _, c := range controls {
		if _, err := tx.Exec(`INSERT INTO protection_baseline_controls (record_id,item_no,domain_name,control_category,requirement_level,control_description,measure_notes,applies,implementation_notes,compliance,finding,remarks) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
			id, c.ItemNo, c.DomainName, c.ControlCategory, c.RequirementLevel, c.ControlDescription, c.MeasureNotes, c.Applies, c.ImplementationNotes, c.Compliance, c.Finding, c.Remarks); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (d *DB) UpdateProtectionBaselineRecord(r *models.ProtectionBaselineRecord) error {
	tx, err := d.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE protection_baseline_records SET system_name=?,security_level=?,filled_by=?,form_date=?,reviewer=?,review_date=?,status=?,creator=?,remarks=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		r.SystemName, r.SecurityLevel, r.FilledBy, r.FormDate, r.Reviewer, r.ReviewDate, r.Status, r.Creator, r.Remarks, r.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM protection_baseline_controls WHERE record_id=?`, r.ID); err != nil {
		return err
	}
	for _, c := range r.Controls {
		if _, err := tx.Exec(`INSERT INTO protection_baseline_controls (record_id,item_no,domain_name,control_category,requirement_level,control_description,measure_notes,applies,implementation_notes,compliance,finding,remarks) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
			r.ID, c.ItemNo, c.DomainName, c.ControlCategory, c.RequirementLevel, c.ControlDescription, c.MeasureNotes, c.Applies, c.ImplementationNotes, c.Compliance, c.Finding, c.Remarks); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (d *DB) DeleteProtectionBaselineRecord(id int) error {
	return d.DeleteProtectionBaselineRecordByCreator(id, "")
}

func (d *DB) DeleteProtectionBaselineRecordByCreator(id int, creator string) error {
	query := `DELETE FROM protection_baseline_records WHERE id=?`
	args := []interface{}{id}
	if strings.TrimSpace(creator) != "" {
		query += ` AND creator=?`
		args = append(args, creator)
	}
	_, err := d.conn.Exec(query, args...)
	return err
}
