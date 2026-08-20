package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ballast-watch/internal/model"
)

// ComplianceRuleStore 告警规则存储。
type ComplianceRuleStore interface {
	Create(ctx context.Context, r *model.ComplianceRule) error
	GetByID(ctx context.Context, id int64) (*model.ComplianceRule, error)
	GetByCode(ctx context.Context, code string) (*model.ComplianceRule, error)
	ListByBallastTank(ctx context.Context, tankID int64) ([]*model.ComplianceRule, error)
	ListEnabled(ctx context.Context) ([]*model.ComplianceRule, error)
	ListAll(ctx context.Context) ([]*model.ComplianceRule, error)
	Update(ctx context.Context, r *model.ComplianceRule) error
	SetEnabled(ctx context.Context, id int64, enabled bool) error
}

type SQLComplianceRuleStore struct {
	db *DB
}

func NewComplianceRuleStore(db *DB) ComplianceRuleStore {
	return &SQLComplianceRuleStore{db: db}
}

const ruleCols = "id, tank_id, code, param_type, op, threshold, duration_sec, level, enabled, created_at"

func scanRule(row interface{ Scan(...any) error }) (*model.ComplianceRule, error) {
	var r model.ComplianceRule
	var createdAt time.Time
	var enabled int
	if err := row.Scan(&r.ID, &r.BallastTankID, &r.Code, &r.ParamType, &r.Op, &r.Threshold,
		&r.DurationSec, &r.Level, &enabled, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	r.Enabled = enabled == 1
	r.CreatedAt = createdAt
	return &r, nil
}

func (s *SQLComplianceRuleStore) Create(ctx context.Context, r *model.ComplianceRule) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO compliance_rules (tank_id, code, param_type, op, threshold, duration_sec, level, enabled, created_at) VALUES (?,?,?,?,?,?,?,?,?)`,
		r.BallastTankID, r.Code, r.ParamType, r.Op, r.Threshold, r.DurationSec, r.Level, boolToInt(r.Enabled), r.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrDuplicateCode
		}
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	r.ID = id
	return nil
}

func (s *SQLComplianceRuleStore) GetByID(ctx context.Context, id int64) (*model.ComplianceRule, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+ruleCols+" FROM compliance_rules WHERE id = ?", id)
	return scanRule(row)
}

func (s *SQLComplianceRuleStore) GetByCode(ctx context.Context, code string) (*model.ComplianceRule, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+ruleCols+" FROM compliance_rules WHERE code = ?", code)
	return scanRule(row)
}

func (s *SQLComplianceRuleStore) ListByBallastTank(ctx context.Context, tankID int64) ([]*model.ComplianceRule, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+ruleCols+" FROM compliance_rules WHERE tank_id=? ORDER BY id", tankID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRules(rows)
}

func (s *SQLComplianceRuleStore) ListEnabled(ctx context.Context) ([]*model.ComplianceRule, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+ruleCols+" FROM compliance_rules WHERE enabled=1 ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRules(rows)
}

func (s *SQLComplianceRuleStore) ListAll(ctx context.Context) ([]*model.ComplianceRule, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+ruleCols+" FROM compliance_rules ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRules(rows)
}

func (s *SQLComplianceRuleStore) Update(ctx context.Context, r *model.ComplianceRule) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE compliance_rules SET code=?, param_type=?, op=?, threshold=?, duration_sec=?, level=?, enabled=? WHERE id=?`,
		r.Code, r.ParamType, r.Op, r.Threshold, r.DurationSec, r.Level, boolToInt(r.Enabled), r.ID)
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

func (s *SQLComplianceRuleStore) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	res, err := s.db.ExecContext(ctx, "UPDATE compliance_rules SET enabled=? WHERE id=?", boolToInt(enabled), id)
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

func scanRules(rows *sql.Rows) ([]*model.ComplianceRule, error) {
	var out []*model.ComplianceRule
	for rows.Next() {
		r, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}