package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ballast-watch/internal/model"
)

// WaterReadingQuery 读数查询条件。
type WaterReadingQuery struct {
	BallastTankID    int64
	SamplingPointID   int64
	ParamType string
	From      time.Time
	To        time.Time
	Limit     int
}

// WaterWaterReadingStore 读数存储。
type WaterWaterReadingStore interface {
	Insert(ctx context.Context, r *model.WaterReading) error
	InsertTreatmentCycle(ctx context.Context, rs []*model.WaterReading) error
	GetByID(ctx context.Context, id int64) (*model.WaterReading, error)
	Query(ctx context.Context, q WaterReadingQuery) ([]*model.WaterReading, error)
	QueryValues(ctx context.Context, sampling_pointID int64, paramType string, from, to time.Time) ([]float64, error)
	QueryValuesWithTime(ctx context.Context, sampling_pointID int64, paramType string, from, to time.Time) ([]model.WaterReading, error)
	ExistsDup(ctx context.Context, sampling_pointID int64, measuredAt time.Time) (bool, error)
	Stats(ctx context.Context, paramType string, from, to time.Time) (*model.WaterWaterReadingStats, error)
	LatestByPoint(ctx context.Context, sampling_pointID int64) (*model.WaterReading, error)
	CountRange(ctx context.Context, tankID int64, from, to time.Time) (int, error)
}

type SQLWaterWaterReadingStore struct {
	db *DB
}

func NewWaterWaterReadingStore(db *DB) WaterWaterReadingStore {
	return &SQLWaterWaterReadingStore{db: db}
}

const water_readingCols = "id, sampling_sampling_point_id, sensor_id, param_type, value, measured_at, raw, created_at"

func scanWaterReading(row interface{ Scan(...any) error }) (*model.WaterReading, error) {
	var r model.WaterReading
	var measured, createdAt time.Time
	var raw sql.NullString
	if err := row.Scan(&r.ID, &r.SamplingPointID, &r.SensorID, &r.ParamType, &r.Value, &measured, &raw, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	r.MeasuredAt = measured
	r.Raw = raw.String
	r.CreatedAt = createdAt
	return &r, nil
}

func (s *SQLWaterWaterReadingStore) Insert(ctx context.Context, r *model.WaterReading) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO water_readings (sampling_sampling_point_id, sensor_id, param_type, value, measured_at, raw, created_at) VALUES (?,?,?,?,?,?,?)`,
		r.SamplingPointID, r.SensorID, r.ParamType, r.Value, r.MeasuredAt, r.Raw, r.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	r.ID = id
	return nil
}

func (s *SQLWaterWaterReadingStore) InsertTreatmentCycle(ctx context.Context, rs []*model.WaterReading) error {
	return s.db.WithTx(ctx, func(tx *sql.Tx) error {
		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO water_readings (sampling_sampling_point_id, sensor_id, param_type, value, measured_at, raw, created_at) VALUES (?,?,?,?,?,?,?)`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		for _, r := range rs {
			if _, err := stmt.ExecContext(ctx, r.SamplingPointID, r.SensorID, r.ParamType, r.Value, r.MeasuredAt, r.Raw, r.CreatedAt); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SQLWaterWaterReadingStore) GetByID(ctx context.Context, id int64) (*model.WaterReading, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+water_readingCols+" FROM water_readings WHERE id = ?", id)
	return scanWaterReading(row)
}

func (s *SQLWaterWaterReadingStore) Query(ctx context.Context, q WaterReadingQuery) ([]*model.WaterReading, error) {
	qry := "SELECT " + water_readingCols + " FROM water_readings WHERE 1=1"
	args := []any{}
	if q.SamplingPointID > 0 {
		qry += " AND sampling_sampling_point_id = ?"
		args = append(args, q.SamplingPointID)
	}
	if q.ParamType != "" {
		qry += " AND param_type = ?"
		args = append(args, q.ParamType)
	}
	if !q.From.IsZero() {
		qry += " AND measured_at >= ?"
		args = append(args, q.From)
	}
	if !q.To.IsZero() {
		qry += " AND measured_at <= ?"
		args = append(args, q.To)
	}
	qry += " ORDER BY measured_at DESC"
	if q.Limit <= 0 || q.Limit > 1000 {
		q.Limit = 200
	}
	qry += " LIMIT ?"
	args = append(args, q.Limit)
	rows, err := s.db.QueryContext(ctx, qry, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.WaterReading
	for rows.Next() {
		r, err := scanWaterReading(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *SQLWaterWaterReadingStore) QueryValues(ctx context.Context, sampling_pointID int64, paramType string, from, to time.Time) ([]float64, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT value FROM water_readings WHERE sampling_sampling_point_id=? AND param_type=? AND measured_at>=? AND measured_at<=? ORDER BY measured_at`,
		sampling_pointID, paramType, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []float64
	for rows.Next() {
		var v float64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *SQLWaterWaterReadingStore) QueryValuesWithTime(ctx context.Context, sampling_pointID int64, paramType string, from, to time.Time) ([]model.WaterReading, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+water_readingCols+` FROM water_readings WHERE sampling_sampling_point_id=? AND param_type=? AND measured_at>=? AND measured_at<=? ORDER BY measured_at`,
		sampling_pointID, paramType, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.WaterReading
	for rows.Next() {
		r, err := scanWaterReading(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *r)
	}
	return out, rows.Err()
}

func (s *SQLWaterWaterReadingStore) ExistsDup(ctx context.Context, sampling_pointID int64, measuredAt time.Time) (bool, error) {
	// 按秒判重：measuredAt 已由上层截断到秒，查询 [measuredAt, measuredAt+1s) 区间内是否已有记录，
	// 这样即便历史数据残留亚秒精度，也能正确识别同一秒内的重复样本。
	var n int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM water_readings WHERE sampling_sampling_point_id=? AND measured_at>=? AND measured_at<?",
		sampling_pointID, measuredAt, measuredAt.Add(time.Second)).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *SQLWaterWaterReadingStore) Stats(ctx context.Context, paramType string, from, to time.Time) (*model.WaterWaterReadingStats, error) {
	st := &model.WaterWaterReadingStats{ParamType: paramType, Unit: model.ParamUnits[paramType]}
	var minV, maxV, avgV float64
	var count, oor int
	row := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*), MIN(value), MAX(value), AVG(value) FROM water_readings WHERE param_type=? AND measured_at>=? AND measured_at<=?`,
		paramType, from, to)
	if err := row.Scan(&count, &minV, &maxV, &avgV); err != nil {
		if errors.Is(err, sql.ErrNoRows) || count == 0 {
			return st, nil
		}
		return nil, err
	}
	st.Count = count
	st.Min = minV
	st.Max = maxV
	st.Avg = avgV
	if count == 0 {
		return st, nil
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT value FROM water_readings WHERE param_type=? AND measured_at>=? AND measured_at<=?`, paramType, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sampling_points []*model.SamplingPoint
	// 该参数所有点位的阈值下界上界
	// 简化：占位，调用方负责 OOR 判定
	_ = sampling_points
	var vals []float64
	for rows.Next() {
		var v float64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		vals = append(vals, v)
	}
	_ = vals
	_ = oor
	return st, rows.Err()
}

func (s *SQLWaterWaterReadingStore) LatestByPoint(ctx context.Context, sampling_pointID int64) (*model.WaterReading, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+water_readingCols+" FROM water_readings WHERE sampling_sampling_point_id=? ORDER BY measured_at DESC LIMIT 1", sampling_pointID)
	return scanWaterReading(row)
}

func (s *SQLWaterWaterReadingStore) CountRange(ctx context.Context, tankID int64, from, to time.Time) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM water_readings r JOIN sampling_sampling_points p ON r.sampling_sampling_point_id=p.id WHERE p.tank_id=? AND r.measured_at>=? AND r.measured_at<=?`,
		tankID, from, to).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}