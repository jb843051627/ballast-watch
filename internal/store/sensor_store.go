package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ballast-watch/internal/model"
)

// SensorStore 传感器存储。
type SensorStore interface {
	Create(ctx context.Context, s *model.Sensor) error
	GetByID(ctx context.Context, id int64) (*model.Sensor, error)
	GetBySerial(ctx context.Context, serial string) (*model.Sensor, error)
	ListByPoint(ctx context.Context, sampling_pointID int64) ([]*model.Sensor, error)
	List(ctx context.Context, limit, offset int) ([]*model.Sensor, error)
	UpdateLastSeen(ctx context.Context, id int64, at time.Time) error
	Update(ctx context.Context, s *model.Sensor) error
	MarkFault(ctx context.Context, id int64) error
	CountByPoint(ctx context.Context, sampling_pointID int64) (int, error)
}

type SQLSensorStore struct {
	db *DB
}

func NewSensorStore(db *DB) SensorStore {
	return &SQLSensorStore{db: db}
}

const sensorCols = "id, sampling_sampling_point_id, serial, vendor, last_seen_at, battery, status, calibration_due_at, created_at"

func scanSensor(row interface{ Scan(...any) error }) (*model.Sensor, error) {
	var s model.Sensor
	var seenAt, calDue, createdAt sql.NullTime
	if err := row.Scan(&s.ID, &s.SamplingPointID, &s.Serial, &s.Vendor, &seenAt, &s.Battery,
		&s.Status, &calDue, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	s.LastSeenAt = seenAt.Time
	s.CalibrationDueAt = calDue.Time
	s.CreatedAt = createdAt.Time
	return &s, nil
}

func (s *SQLSensorStore) Create(ctx context.Context, se *model.Sensor) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO sensors (sampling_sampling_point_id, serial, vendor, last_seen_at, battery, status, calibration_due_at, created_at) VALUES (?,?,?,?,?,?,?,?)`,
		se.SamplingPointID, se.Serial, se.Vendor, se.LastSeenAt, se.Battery, se.Status, se.CalibrationDueAt, se.CreatedAt)
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
	se.ID = id
	return nil
}

func (s *SQLSensorStore) GetByID(ctx context.Context, id int64) (*model.Sensor, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+sensorCols+" FROM sensors WHERE id = ?", id)
	return scanSensor(row)
}

func (s *SQLSensorStore) GetBySerial(ctx context.Context, serial string) (*model.Sensor, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+sensorCols+" FROM sensors WHERE serial = ?", serial)
	return scanSensor(row)
}

func (s *SQLSensorStore) ListByPoint(ctx context.Context, sampling_pointID int64) ([]*model.Sensor, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+sensorCols+" FROM sensors WHERE sampling_sampling_point_id = ? ORDER BY id", sampling_pointID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Sensor
	for rows.Next() {
		se, err := scanSensor(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, se)
	}
	return out, rows.Err()
}

func (s *SQLSensorStore) List(ctx context.Context, limit, offset int) ([]*model.Sensor, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+sensorCols+" FROM sensors ORDER BY id LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Sensor
	for rows.Next() {
		se, err := scanSensor(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, se)
	}
	return out, rows.Err()
}

func (s *SQLSensorStore) UpdateLastSeen(ctx context.Context, id int64, at time.Time) error {
	_, err := s.db.ExecContext(ctx, "UPDATE sensors SET last_seen_at=?, status='active' WHERE id=?", at, id)
	return err
}

func (s *SQLSensorStore) Update(ctx context.Context, se *model.Sensor) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE sensors SET sampling_sampling_point_id=?, serial=?, vendor=?, battery=?, status=?, calibration_due_at=? WHERE id=?`,
		se.SamplingPointID, se.Serial, se.Vendor, se.Battery, se.Status, se.CalibrationDueAt, se.ID)
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

func (s *SQLSensorStore) MarkFault(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, "UPDATE sensors SET status='fault' WHERE id=?", id)
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

func (s *SQLSensorStore) CountByPoint(ctx context.Context, sampling_pointID int64) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sensors WHERE sampling_sampling_point_id=?", sampling_pointID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}