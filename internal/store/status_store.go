package store

import (
	"context"
	"database/sql"
	"errors"

	"ballast-watch/internal/model"
)

// ComplianceStatusStore 洁净状态流转记录存储。
type ComplianceStatusStore interface {
	Append(ctx context.Context, s *model.ComplianceStatus) error
	ListByBallastTank(ctx context.Context, tankID int64, limit int) ([]*model.ComplianceStatus, error)
	LatestByBallastTank(ctx context.Context, tankID int64) (*model.ComplianceStatus, error)
	CountByState(ctx context.Context, state string) (int, error)
}

type SQLComplianceStatusStore struct {
	db *DB
}

func NewComplianceStatusStore(db *DB) ComplianceStatusStore {
	return &SQLComplianceStatusStore{db: db}
}

const statusCols = "id, tank_id, cycle_id, state, reason, changed_at"

func scanStatus(row interface{ Scan(...any) error }) (*model.ComplianceStatus, error) {
	var s model.ComplianceStatus
	var reason sql.NullString
	if err := row.Scan(&s.ID, &s.BallastTankID, &s.TreatmentCycleID, &s.State, &reason, &s.ChangedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	s.Reason = reason.String
	return &s, nil
}

func (s *SQLComplianceStatusStore) Append(ctx context.Context, st *model.ComplianceStatus) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO compliance_status (tank_id, cycle_id, state, reason, changed_at) VALUES (?,?,?,?,?)`,
		st.BallastTankID, st.TreatmentCycleID, st.State, st.Reason, st.ChangedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	st.ID = id
	return nil
}

func (s *SQLComplianceStatusStore) ListByBallastTank(ctx context.Context, tankID int64, limit int) ([]*model.ComplianceStatus, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, "SELECT "+statusCols+" FROM compliance_status WHERE tank_id=? ORDER BY changed_at DESC LIMIT ?", tankID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.ComplianceStatus
	for rows.Next() {
		st, err := scanStatus(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

func (s *SQLComplianceStatusStore) LatestByBallastTank(ctx context.Context, tankID int64) (*model.ComplianceStatus, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+statusCols+" FROM compliance_status WHERE tank_id=? ORDER BY changed_at DESC LIMIT 1", tankID)
	return scanStatus(row)
}

func (s *SQLComplianceStatusStore) CountByState(ctx context.Context, state string) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM compliance_status WHERE state=?", state).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}