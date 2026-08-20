package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"ballast-watch/internal/model"
)

// VesselStore 洁净区存储。
type VesselStore interface {
	Create(ctx context.Context, c *model.Vessel) error
	GetByID(ctx context.Context, id int64) (*model.Vessel, error)
	GetByCode(ctx context.Context, code string) (*model.Vessel, error)
	List(ctx context.Context, limit, offset int) ([]*model.Vessel, error)
	Update(ctx context.Context, c *model.Vessel) error
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int, error)
}

type SQLVesselStore struct {
	db *DB
}

func NewVesselStore(db *DB) VesselStore {
	return &SQLVesselStore{db: db}
}

const vesselCols = "id, name, code, grade, area_sqm, status, created_at"

func scanVessel(row *sql.Row) (*model.Vessel, error) {
	var c model.Vessel
	var createdAt time.Time
	if err := row.Scan(&c.ID, &c.Name, &c.Code, &c.Grade, &c.AreaSqm, &c.Status, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	c.CreatedAt = createdAt
	return &c, nil
}

func (s *SQLVesselStore) Create(ctx context.Context, c *model.Vessel) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO vessels (name, code, grade, area_sqm, status, created_at) VALUES (?,?,?,?,?,?)`,
		c.Name, c.Code, c.Grade, c.AreaSqm, c.Status, c.CreatedAt)
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
	c.ID = id
	return nil
}

func (s *SQLVesselStore) GetByID(ctx context.Context, id int64) (*model.Vessel, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+vesselCols+" FROM vessels WHERE id = ?", id)
	return scanVessel(row)
}

func (s *SQLVesselStore) GetByCode(ctx context.Context, code string) (*model.Vessel, error) {
	row := s.db.QueryRowContext(ctx, "SELECT "+vesselCols+" FROM vessels WHERE code = ?", code)
	return scanVessel(row)
}

func (s *SQLVesselStore) List(ctx context.Context, limit, offset int) ([]*model.Vessel, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT "+vesselCols+" FROM vessels ORDER BY id LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Vessel
	for rows.Next() {
		var c model.Vessel
		var createdAt time.Time
		if err := rows.Scan(&c.ID, &c.Name, &c.Code, &c.Grade, &c.AreaSqm, &c.Status, &createdAt); err != nil {
			return nil, err
		}
		c.CreatedAt = createdAt
		out = append(out, &c)
	}
	return out, rows.Err()
}

func (s *SQLVesselStore) Update(ctx context.Context, c *model.Vessel) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE vessels SET name=?, code=?, grade=?, area_sqm=? WHERE id=?`,
		c.Name, c.Code, c.Grade, c.AreaSqm, c.ID)
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

func (s *SQLVesselStore) Delete(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, "DELETE FROM vessels WHERE id = ?", id)
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

func (s *SQLVesselStore) Count(ctx context.Context) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vessels").Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}