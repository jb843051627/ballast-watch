package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ballast-watch/internal/model"
)

// ComplianceComplianceAlertStore 告警存储。
type ComplianceComplianceAlertStore interface {
	Create(ctx context.Context, a *model.ComplianceAlert) error
	GetByID(ctx context.Context, id int64) (*model.ComplianceAlert, error)
	List(ctx context.Context, in model.ComplianceComplianceAlertInput) ([]*model.ComplianceAlert, error)
	ListOpenByBallastTank(ctx context.Context, tankID int64) ([]*model.ComplianceAlert, error)
	OpenByRulePoint(ctx context.Context, ruleID, sampling_pointID int64) (*model.ComplianceAlert, error)
	SetAck(ctx context.Context, id int64, at time.Time) error
	SetResolved(ctx context.Context, id int64, at time.Time) error
	CountOpenByBallastTank(ctx context.Context, tankID int64) (int, error)
	CountOpenByLevel(ctx context.Context, tankID int64, level string) (int, error)
}

type SQLComplianceComplianceAlertStore struct {
	db *DB
}

func NewComplianceComplianceAlertStore(db *DB) ComplianceComplianceAlertStore {
	return &SQLComplianceComplianceAlertStore{db: db}
}

const compliance_alertCols = "id, rule_id, tank_id, sampling_sampling_point_id, level, message, status, opened_at, ack_at, resolved_at"

func scanComplianceAlert(row interface{ Scan(...any) error }) (*model.ComplianceAlert, error) {
	var a model.ComplianceAlert
	var ack, resolved sql.NullTime
	var msg sql.NullString
	if err := row.Scan(&a.ID, &a.RuleID, &a.BallastTankID, &a.SamplingPointID, &a.Level, &msg, &a.Status,
		&a.OpenedAt, &ack, &resolved); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	a.Message = msg.String
	if ack.Valid {
		t := ack.Time
		a.AckAt = &t
	}
	if resolved.Valid {
		t := resolved.Time
		a.ResolvedAt = &t
	}
	return &a, nil
}

func (s *SQLComplianceComplianceAlertStore) Create(ctx context.Context, a *model.ComplianceAlert) error {
	var ack, resolved sql.NullTime
	if a.AckAt != nil {
		ack.Time = *a.AckAt
		ack.Valid = true
	}
	if a.ResolvedAt != nil {
		resolved.Time = *a.ResolvedAt
		resolved.Valid = true
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO compliance_compliance_alerts (rule_id, tank_id, sampling_sampling_point_id, level, message, status, opened_at, ack_at, resolved_at) VALUES (?,?,?,?,?,?,?,?,?)`,
		a.RuleID, a.BallastTankID, a.SamplingPointID, a.Level, a.Message, a.Status, a.OpenedAt, ack, resolved)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	a.ID = id
	return nil
}

func (s *SQLComplianceComplianceAlertStore) GetByID(ctx context.Context, id int64) (*model.ComplianceAlert, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+compliance_alertCols+" FROM compliance_compliance_alerts WHERE id = ?", id)
	return scanComplianceAlert(row)
}

func (s *SQLComplianceComplianceAlertStore) List(ctx context.Context, in model.ComplianceComplianceAlertInput) ([]*model.ComplianceAlert, error) {
	q := "SELECT " + compliance_alertCols + " FROM compliance_compliance_alerts WHERE 1=1"
	args := []any{}
	if in.BallastTankID > 0 {
		q += " AND tank_id = ?"
		args = append(args, in.BallastTankID)
	}
	if in.Level != "" {
		q += " AND level = ?"
		args = append(args, in.Level)
	}
	if in.Status != "" {
		q += " AND status = ?"
		args = append(args, in.Status)
	}
	if !in.From.IsZero() {
		q += " AND opened_at >= ?"
		args = append(args, in.From)
	}
	if !in.To.IsZero() {
		q += " AND opened_at <= ?"
		args = append(args, in.To)
	}
	q += " ORDER BY opened_at DESC"
	if in.Limit <= 0 || in.Limit > 500 {
		in.Limit = 100
	}
	q += " LIMIT ?"
	args = append(args, in.Limit)
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.ComplianceAlert
	for rows.Next() {
		a, err := scanComplianceAlert(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *SQLComplianceComplianceAlertStore) ListOpenByBallastTank(ctx context.Context, tankID int64) ([]*model.ComplianceAlert, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT "+compliance_alertCols+" FROM compliance_compliance_alerts WHERE tank_id=? AND status IN ('open','acknowledged') ORDER BY opened_at",
		tankID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.ComplianceAlert
	for rows.Next() {
		a, err := scanComplianceAlert(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *SQLComplianceComplianceAlertStore) OpenByRulePoint(ctx context.Context, ruleID, sampling_pointID int64) (*model.ComplianceAlert, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT "+compliance_alertCols+" FROM compliance_compliance_alerts WHERE rule_id=? AND sampling_sampling_point_id=? AND status IN ('open','acknowledged') ORDER BY opened_at LIMIT 1",
		ruleID, sampling_pointID)
	a, err := scanComplianceAlert(row)
	if err != nil { return nil, fmt.Errorf("open alert: %w", err) }
	return a, nil
}

func (s *SQLComplianceComplianceAlertStore) SetAck(ctx context.Context, id int64, at time.Time) error {
	res, err := s.db.ExecContext(ctx, "UPDATE compliance_compliance_alerts SET status='acknowledged', ack_at=? WHERE id=?", at, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (s *SQLComplianceComplianceAlertStore) SetResolved(ctx context.Context, id int64, at time.Time) error {
	res, err := s.db.ExecContext(ctx, "UPDATE compliance_compliance_alerts SET status='resolved', resolved_at=? WHERE id=?", at, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (s *SQLComplianceComplianceAlertStore) CountOpenByBallastTank(ctx context.Context, tankID int64) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM compliance_compliance_alerts WHERE tank_id=? AND status IN ('open','acknowledged')", tankID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (s *SQLComplianceComplianceAlertStore) CountOpenByLevel(ctx context.Context, tankID int64, level string) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM compliance_compliance_alerts WHERE tank_id=? AND status IN ('open','acknowledged') AND level=?", tankID, level).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}