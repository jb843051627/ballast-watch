package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ballast-watch/internal/model"
)

// TreatmentCycleStore 批次存储。
type TreatmentCycleStore interface {
	Create(ctx context.Context, b *model.TreatmentCycle) error
	GetByID(ctx context.Context, id int64) (*model.TreatmentCycle, error)
	GetActiveByBallastTank(ctx context.Context, tankID int64) (*model.TreatmentCycle, error)
	ListByBallastTank(ctx context.Context, tankID int64) ([]*model.TreatmentCycle, error)
	List(ctx context.Context, limit, offset int) ([]*model.TreatmentCycle, error)
	UpdateStatus(ctx context.Context, id int64, status string, endAt *time.Time) error
	CountActiveByBallastTank(ctx context.Context, tankID int64) (int, error)
}

type SQLTreatmentCycleStore struct {
	db *DB
}

func NewTreatmentCycleStore(db *DB) TreatmentCycleStore {
	return &SQLTreatmentCycleStore{db: db}
}

const cycleCols = "id, tank_id, name, product, phase, start_at, end_at, status, created_at"

func scanTreatmentCycle(row interface{ Scan(...any) error }) (*model.TreatmentCycle, error) {
	var b model.TreatmentCycle
	var endAt, createdAt sql.NullTime
	if err := row.Scan(&b.ID, &b.BallastTankID, &b.Name, &b.Product, &b.Phase, &b.StartAt,
		&endAt, &b.Status, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	if endAt.Valid {
		t := endAt.Time
		b.EndAt = &t
	}
	b.CreatedAt = createdAt.Time
	return &b, nil
}

func (s *SQLTreatmentCycleStore) Create(ctx context.Context, b *model.TreatmentCycle) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO treatment_cycles (tank_id, name, product, phase, start_at, end_at, status, created_at) VALUES (?,?,?,?,?,?,?,?)`,
		b.BallastTankID, b.Name, b.Product, b.Phase, b.StartAt, b.EndAt, b.Status, b.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	b.ID = id
	return nil
}

func (s *SQLTreatmentCycleStore) GetByID(ctx context.Context, id int64) (*model.TreatmentCycle, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+cycleCols+" FROM treatment_cycles WHERE id = ?", id)
	return scanTreatmentCycle(row)
}

func (s *SQLTreatmentCycleStore) GetActiveByBallastTank(ctx context.Context, tankID int64) (*model.TreatmentCycle, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT "+cycleCols+" FROM treatment_cycles WHERE tank_id=? AND status IN ('planning','in_progress') ORDER BY id DESC LIMIT 1", tankID)
	return scanTreatmentCycle(row)
}

func (s *SQLTreatmentCycleStore) ListByBallastTank(ctx context.Context, tankID int64) ([]*model.TreatmentCycle, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+cycleCols+" FROM treatment_cycles WHERE tank_id=? ORDER BY start_at DESC", tankID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.TreatmentCycle
	for rows.Next() {
		b, err := scanTreatmentCycle(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *SQLTreatmentCycleStore) List(ctx context.Context, limit, offset int) ([]*model.TreatmentCycle, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+cycleCols+" FROM treatment_cycles ORDER BY id DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.TreatmentCycle
	for rows.Next() {
		b, err := scanTreatmentCycle(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *SQLTreatmentCycleStore) UpdateStatus(ctx context.Context, id int64, status string, endAt *time.Time) error {
	res, err := s.db.ExecContext(ctx, "UPDATE treatment_cycles SET status=?, end_at=? WHERE id=?", status, endAt, id)
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

func (s *SQLTreatmentCycleStore) CountActiveByBallastTank(ctx context.Context, tankID int64) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM treatment_cycles WHERE tank_id=? AND status IN ('planning','in_progress')", tankID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}