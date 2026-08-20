package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ballast-watch/internal/model"
)

// CalibrationStore 校准记录存储。
type CalibrationStore interface {
	Create(ctx context.Context, c *model.Calibration) error
	GetByID(ctx context.Context, id int64) (*model.Calibration, error)
	ListBySensor(ctx context.Context, sensorID int64, limit int) ([]*model.Calibration, error)
	LatestBySensor(ctx context.Context, sensorID int64) (*model.Calibration, error)
	CountDueSensors(ctx context.Context, now time.Time) (int, error)
}

type SQLCalibrationStore struct {
	db *DB
}

func NewCalibrationStore(db *DB) CalibrationStore {
	return &SQLCalibrationStore{db: db}
}

const calCols = "id, sensor_id, performed_at, due_at, standard, result, operator, created_at"

func scanCal(row interface{ Scan(...any) error }) (*model.Calibration, error) {
	var c model.Calibration
	var createdAt time.Time
	if err := row.Scan(&c.ID, &c.SensorID, &c.PerformedAt, &c.DueAt, &c.Standard, &c.Result,
		&c.Operator, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	c.CreatedAt = createdAt
	return &c, nil
}

func (s *SQLCalibrationStore) Create(ctx context.Context, c *model.Calibration) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO calibrations (sensor_id, performed_at, due_at, standard, result, operator, created_at) VALUES (?,?,?,?,?,?,?)`,
		c.SensorID, c.PerformedAt, c.DueAt, c.Standard, c.Result, c.Operator, c.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	c.ID = id
	return nil
}

func (s *SQLCalibrationStore) GetByID(ctx context.Context, id int64) (*model.Calibration, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+calCols+" FROM calibrations WHERE id = ?", id)
	return scanCal(row)
}

func (s *SQLCalibrationStore) ListBySensor(ctx context.Context, sensorID int64, limit int) ([]*model.Calibration, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, "SELECT "+calCols+" FROM calibrations WHERE sensor_id=? ORDER BY performed_at DESC LIMIT ?", sensorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Calibration
	for rows.Next() {
		c, err := scanCal(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *SQLCalibrationStore) LatestBySensor(ctx context.Context, sensorID int64) (*model.Calibration, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+calCols+" FROM calibrations WHERE sensor_id=? ORDER BY performed_at DESC LIMIT 1", sensorID)
	return scanCal(row)
}

func (s *SQLCalibrationStore) CountDueSensors(ctx context.Context, now time.Time) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM sensors WHERE calibration_due_at IS NOT NULL AND calibration_due_at <= ? AND status != 'fault'", now).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}