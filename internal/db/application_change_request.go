package db

import (
	"isms-privilege/internal/models"
	"strings"
)

const applicationChangeRequestColumns = `id,suggestor,form_date,approver,related_system,system_name,feature_name,is_required_feature,is_major_impact,expected_online_date,background_description,existing_security_measures,security_scope_involved,access_control_measures,audit_measures,continuity_measures,identification_measures,system_acquisition_measures,communication_measures,integrity_measures,capacity_management,capacity_change_description,other_security_measures,other_security_description,information_service_opinion,meeting_time,meeting_decision,rejection_reason,coordinating_staff,coordination_date,coordination_approver,status,creator,remarks,created_at,updated_at`

func scanApplicationChangeRequest(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.ApplicationChangeRequest, error) {
	var r models.ApplicationChangeRequest
	err := scanner.Scan(
		&r.ID, &r.Suggestor, &r.FormDate, &r.Approver, &r.RelatedSystem, &r.SystemName, &r.FeatureName,
		&r.IsRequiredFeature, &r.IsMajorImpact, &r.ExpectedOnlineDate, &r.BackgroundDescription,
		&r.ExistingSecurityMeasures, &r.SecurityScopeInvolved, &r.AccessControlMeasures, &r.AuditMeasures,
		&r.ContinuityMeasures, &r.IdentificationMeasures, &r.SystemAcquisitionMeasures,
		&r.CommunicationMeasures, &r.IntegrityMeasures, &r.CapacityManagement,
		&r.CapacityChangeDescription, &r.OtherSecurityMeasures, &r.OtherSecurityDescription,
		&r.InformationServiceOpinion, &r.MeetingTime, &r.MeetingDecision, &r.RejectionReason,
		&r.CoordinatingStaff, &r.CoordinationDate, &r.CoordinationApprover, &r.Status, &r.Creator,
		&r.Remarks, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (d *DB) ListApplicationChangeRequests() ([]models.ApplicationChangeRequest, error) {
	return d.ListApplicationChangeRequestsByCreator("")
}

func (d *DB) ListApplicationChangeRequestsByCreator(creator string) ([]models.ApplicationChangeRequest, error) {
	query := `SELECT ` + applicationChangeRequestColumns + ` FROM application_change_requests`
	args := []interface{}{}
	if strings.TrimSpace(creator) != "" {
		query += ` WHERE creator=?`
		args = append(args, creator)
	}
	query += ` ORDER BY id DESC`
	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.ApplicationChangeRequest{}
	for rows.Next() {
		r, err := scanApplicationChangeRequest(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *r)
	}
	return list, rows.Err()
}

func (d *DB) GetApplicationChangeRequest(id int) (*models.ApplicationChangeRequest, error) {
	return d.GetApplicationChangeRequestByCreator(id, "")
}

func (d *DB) GetApplicationChangeRequestByCreator(id int, creator string) (*models.ApplicationChangeRequest, error) {
	query := `SELECT ` + applicationChangeRequestColumns + ` FROM application_change_requests WHERE id=?`
	args := []interface{}{id}
	if strings.TrimSpace(creator) != "" {
		query += ` AND creator=?`
		args = append(args, creator)
	}
	return scanApplicationChangeRequest(d.conn.QueryRow(query, args...))
}

func (d *DB) CreateApplicationChangeRequest(r *models.ApplicationChangeRequest) (int64, error) {
	res, err := d.conn.Exec(`INSERT INTO application_change_requests (suggestor,form_date,approver,related_system,system_name,feature_name,is_required_feature,is_major_impact,expected_online_date,background_description,existing_security_measures,security_scope_involved,access_control_measures,audit_measures,continuity_measures,identification_measures,system_acquisition_measures,communication_measures,integrity_measures,capacity_management,capacity_change_description,other_security_measures,other_security_description,information_service_opinion,meeting_time,meeting_decision,rejection_reason,coordinating_staff,coordination_date,coordination_approver,status,creator,remarks) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		r.Suggestor, r.FormDate, r.Approver, r.RelatedSystem, r.SystemName, r.FeatureName, r.IsRequiredFeature,
		r.IsMajorImpact, r.ExpectedOnlineDate, r.BackgroundDescription, r.ExistingSecurityMeasures,
		r.SecurityScopeInvolved, r.AccessControlMeasures, r.AuditMeasures, r.ContinuityMeasures,
		r.IdentificationMeasures, r.SystemAcquisitionMeasures, r.CommunicationMeasures,
		r.IntegrityMeasures, r.CapacityManagement, r.CapacityChangeDescription,
		r.OtherSecurityMeasures, r.OtherSecurityDescription, r.InformationServiceOpinion,
		r.MeetingTime, r.MeetingDecision, r.RejectionReason, r.CoordinatingStaff,
		r.CoordinationDate, r.CoordinationApprover, r.Status, r.Creator, r.Remarks)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (d *DB) UpdateApplicationChangeRequest(r *models.ApplicationChangeRequest) error {
	_, err := d.conn.Exec(`UPDATE application_change_requests SET suggestor=?,form_date=?,approver=?,related_system=?,system_name=?,feature_name=?,is_required_feature=?,is_major_impact=?,expected_online_date=?,background_description=?,existing_security_measures=?,security_scope_involved=?,access_control_measures=?,audit_measures=?,continuity_measures=?,identification_measures=?,system_acquisition_measures=?,communication_measures=?,integrity_measures=?,capacity_management=?,capacity_change_description=?,other_security_measures=?,other_security_description=?,information_service_opinion=?,meeting_time=?,meeting_decision=?,rejection_reason=?,coordinating_staff=?,coordination_date=?,coordination_approver=?,status=?,creator=?,remarks=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		r.Suggestor, r.FormDate, r.Approver, r.RelatedSystem, r.SystemName, r.FeatureName, r.IsRequiredFeature,
		r.IsMajorImpact, r.ExpectedOnlineDate, r.BackgroundDescription, r.ExistingSecurityMeasures,
		r.SecurityScopeInvolved, r.AccessControlMeasures, r.AuditMeasures, r.ContinuityMeasures,
		r.IdentificationMeasures, r.SystemAcquisitionMeasures, r.CommunicationMeasures,
		r.IntegrityMeasures, r.CapacityManagement, r.CapacityChangeDescription,
		r.OtherSecurityMeasures, r.OtherSecurityDescription, r.InformationServiceOpinion,
		r.MeetingTime, r.MeetingDecision, r.RejectionReason, r.CoordinatingStaff,
		r.CoordinationDate, r.CoordinationApprover, r.Status, r.Creator, r.Remarks, r.ID)
	return err
}

func (d *DB) DeleteApplicationChangeRequest(id int) error {
	return d.DeleteApplicationChangeRequestByCreator(id, "")
}

func (d *DB) DeleteApplicationChangeRequestByCreator(id int, creator string) error {
	query := `DELETE FROM application_change_requests WHERE id=?`
	args := []interface{}{id}
	if strings.TrimSpace(creator) != "" {
		query += ` AND creator=?`
		args = append(args, creator)
	}
	_, err := d.conn.Exec(query, args...)
	return err
}
