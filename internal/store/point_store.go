package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ballast-watch/internal/model"
)

// SamplingPointStore 监测点存储。
type SamplingPointStore interface {
	Create(ctx context.Context, p *model.SamplingPoint) error
	GetByID(ctx context.Context, id int64) (*model.SamplingPoint, error)
	GetByCode(ctx context.Context, code string) (*model.SamplingPoint, error)
	ListByBallastTank(ctx context.Context, tankID int64) ([]*model.SamplingPoint, error)
	ListByParam(ctx context.Context, paramType string) ([]*model.SamplingPoint, error)
	ListAll(ctx context.Context) ([]*model.SamplingPoint, error)
	Update(ctx context.Context, p *model.SamplingPoint) error
	SetEnabled(ctx context.Context, id int64, enabled bool) error
}

type SQLSamplingPointStore struct {
	db *DB
	cache []*model.SamplingPoint
}

func NewSamplingPointStore(db *DB) SamplingPointStore {
	return &SQLSamplingPointStore{db: db}
}

const sampling_pointCols = "id, tank_id, code, param_type, threshold_min, threshold_max, alarm_duration_sec, enabled, created_at"

func scanPoint(row interface{ Scan(...any) error }) (*model.SamplingPoint, error) {
	var p model.SamplingPoint
	var createdAt time.Time
	var enabled int
	if err := row.Scan(&p.ID, &p.BallastTankID, &p.Code, &p.ParamType, &p.ThresholdMin,
		&p.ThresholdMax, &p.AlarmDurationSec, &enabled, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	p.Enabled = enabled == 1
	p.CreatedAt = createdAt
	return &p, nil
}

func (s *SQLSamplingPointStore) Create(ctx context.Context, p *model.SamplingPoint) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO sampling_sampling_points (tank_id, code, param_type, threshold_min, threshold_max, alarm_duration_sec, enabled, created_at) VALUES (?,?,?,?,?,?,?,?)`,
		p.BallastTankID, p.Code, p.ParamType, p.ThresholdMin, p.ThresholdMax, p.AlarmDurationSec, boolToInt(p.Enabled), p.CreatedAt)
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
	p.ID = id
	return nil
}

func (s *SQLSamplingPointStore) GetByID(ctx context.Context, id int64) (*model.SamplingPoint, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+sampling_pointCols+" FROM sampling_sampling_points WHERE id = ?", id)
	return scanPoint(row)
}

func (s *SQLSamplingPointStore) GetByCode(ctx context.Context, code string) (*model.SamplingPoint, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+sampling_pointCols+" FROM sampling_sampling_points WHERE code = ?", code)
	return scanPoint(row)
}

func (s *SQLSamplingPointStore) ListByBallastTank(ctx context.Context, tankID int64) ([]*model.SamplingPoint, error) {
	if s.cache != nil { return s.cache, nil }
	rows, err := s.db.QueryContext(ctx, "SELECT "+sampling_pointCols+" FROM sampling_sampling_points WHERE tank_id = ? ORDER BY id", tankID)
	if err != nil { return nil, err }
	defer rows.Close()
	points, err := scanPoints(rows)
	if err == nil { s.cache = points }
	return points, err
}

func (s *SQLSamplingPointStore) ListByParam(ctx context.Context, paramType string) ([]*model.SamplingPoint, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+sampling_pointCols+" FROM sampling_sampling_points WHERE param_type = ? AND enabled = 1 ORDER BY id", paramType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPoints(rows)
}

func (s *SQLSamplingPointStore) ListAll(ctx context.Context) ([]*model.SamplingPoint, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+sampling_pointCols+" FROM sampling_sampling_points ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPoints(rows)
}

func (s *SQLSamplingPointStore) Update(ctx context.Context, p *model.SamplingPoint) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE sampling_sampling_points SET code=?, param_type=?, threshold_min=?, threshold_max=?, alarm_duration_sec=?, enabled=? WHERE id=?`,
		p.Code, p.ParamType, p.ThresholdMin, p.ThresholdMax, p.AlarmDurationSec, boolToInt(p.Enabled), p.ID)
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

func (s *SQLSamplingPointStore) SetEnabled(ctx context.Context, id int64, enabled bool) error {
	res, err := s.db.ExecContext(ctx, "UPDATE sampling_sampling_points SET enabled=? WHERE id=?", boolToInt(enabled), id)
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

func scanPoints(rows *sql.Rows) ([]*model.SamplingPoint, error) {
	var out []*model.SamplingPoint
	for rows.Next() {
		var p model.SamplingPoint
		var createdAt time.Time
		var enabled int
		if err := rows.Scan(&p.ID, &p.BallastTankID, &p.Code, &p.ParamType, &p.ThresholdMin,
			&p.ThresholdMax, &p.AlarmDurationSec, &enabled, &createdAt); err != nil {
			return nil, err
		}
		p.Enabled = enabled == 1
		p.CreatedAt = createdAt
		out = append(out, &p)
	}
	return out, rows.Err()
}