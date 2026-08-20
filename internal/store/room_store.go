package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ballast-watch/internal/model"
)

// TankStore 房间存储。
type TankStore interface {
	Create(ctx context.Context, r *model.BallastTank) error
	GetByID(ctx context.Context, id int64) (*model.BallastTank, error)
	GetByCode(ctx context.Context, code string) (*model.BallastTank, error)
	ListByVessel(ctx context.Context, vesselID int64) ([]*model.BallastTank, error)
	List(ctx context.Context, limit, offset int) ([]*model.BallastTank, error)
	Update(ctx context.Context, r *model.BallastTank) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	CountByVessel(ctx context.Context, vesselID int64) (int, error)
}

type SQLTankStore struct {
	db *DB
}

func NewTankStore(db *DB) TankStore {
	return &SQLTankStore{db: db}
}

const tankCols = "id, vessel_id, name, code, kind, area_sqm, target_pressure, pressure_tolerance, target_temp, target_humidity, status, created_at"

func scanBallastTank(row interface{ Scan(...any) error }) (*model.BallastTank, error) {
	var r model.BallastTank
	var createdAt time.Time
	if err := row.Scan(&r.ID, &r.VesselID, &r.Name, &r.Code, &r.Kind, &r.AreaSqm,
		&r.TargetPressure, &r.PressureTolerance, &r.TargetTemp, &r.TargetHumidity,
		&r.Status, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	r.CreatedAt = createdAt
	return &r, nil
}

func (s *SQLTankStore) Create(ctx context.Context, r *model.BallastTank) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO tanks (vessel_id, name, code, kind, area_sqm, target_pressure, pressure_tolerance, target_temp, target_humidity, status, created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		r.VesselID, r.Name, r.Code, r.Kind, r.AreaSqm, r.TargetPressure, r.PressureTolerance,
		r.TargetTemp, r.TargetHumidity, r.Status, r.CreatedAt)
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

func (s *SQLTankStore) GetByID(ctx context.Context, id int64) (*model.BallastTank, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+tankCols+" FROM tanks WHERE id = ?", id)
	return scanBallastTank(row)
}

func (s *SQLTankStore) GetByCode(ctx context.Context, code string) (*model.BallastTank, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+tankCols+" FROM tanks WHERE code = ?", code)
	return scanBallastTank(row)
}

func (s *SQLTankStore) ListByVessel(ctx context.Context, vesselID int64) ([]*model.BallastTank, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+tankCols+" FROM tanks WHERE vessel_id = ? ORDER BY id", vesselID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.BallastTank
	for rows.Next() {
		r, err := scanBallastTank(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *SQLTankStore) List(ctx context.Context, limit, offset int) ([]*model.BallastTank, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+tankCols+" FROM tanks ORDER BY id LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.BallastTank
	for rows.Next() {
		r, err := scanBallastTank(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *SQLTankStore) Update(ctx context.Context, r *model.BallastTank) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE tanks SET name=?, code=?, kind=?, area_sqm=?, target_pressure=?, pressure_tolerance=?, target_temp=?, target_humidity=? WHERE id=?`,
		r.Name, r.Code, r.Kind, r.AreaSqm, r.TargetPressure, r.PressureTolerance, r.TargetTemp, r.TargetHumidity, r.ID)
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

func (s *SQLTankStore) UpdateStatus(ctx context.Context, id int64, status string) error {
	res, err := s.db.ExecContext(ctx, "UPDATE tanks SET status=? WHERE id=?", status, id)
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

func (s *SQLTankStore) CountByVessel(ctx context.Context, vesselID int64) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tanks WHERE vessel_id=?", vesselID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}