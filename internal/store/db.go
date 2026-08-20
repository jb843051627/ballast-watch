package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// DB 数据库句柄。
type DB struct {
	*sql.DB
}

// Open 打开 SQLite 数据库（文件落盘，非 :memory:）。
func Open(path string) (*DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", path)
	sqldb, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	sqldb.SetMaxOpenConns(1)
	sqldb.SetMaxIdleConns(1)
	sqldb.SetConnMaxLifetime(time.Hour)
	if err := sqldb.Ping(); err != nil {
		return nil, err
	}
	db := &DB{sqldb}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

// migrate 建表。
func (db *DB) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS vessels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			code TEXT NOT NULL UNIQUE,
			grade TEXT NOT NULL,
			area_sqm REAL NOT NULL,
			status TEXT NOT NULL DEFAULT 'at_rest',
			created_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS tanks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vessel_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			code TEXT NOT NULL UNIQUE,
			kind TEXT NOT NULL,
			area_sqm REAL NOT NULL,
			target_pressure REAL NOT NULL,
			pressure_tolerance REAL NOT NULL,
			target_temp REAL NOT NULL,
			target_humidity REAL NOT NULL,
			status TEXT NOT NULL DEFAULT 'at_rest',
			created_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS sampling_sampling_points (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tank_id INTEGER NOT NULL,
			code TEXT NOT NULL UNIQUE,
			param_type TEXT NOT NULL,
			threshold_min REAL NOT NULL,
			threshold_max REAL NOT NULL,
			alarm_duration_sec INTEGER NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS sensors (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sampling_sampling_point_id INTEGER NOT NULL,
			serial TEXT NOT NULL UNIQUE,
			vendor TEXT NOT NULL,
			last_seen_at DATETIME,
			battery REAL NOT NULL DEFAULT 100,
			status TEXT NOT NULL DEFAULT 'active',
			calibration_due_at DATETIME,
			created_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS water_readings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sampling_sampling_point_id INTEGER NOT NULL,
			sensor_id INTEGER NOT NULL,
			param_type TEXT NOT NULL,
			value REAL NOT NULL,
			measured_at DATETIME NOT NULL,
			raw TEXT,
			created_at DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_water_readings_sampling_point_measured ON water_readings(sampling_sampling_point_id, measured_at)`,
		`CREATE INDEX IF NOT EXISTS idx_water_readings_tank_time ON water_readings(param_type, measured_at)`,
		`CREATE TABLE IF NOT EXISTS compliance_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tank_id INTEGER NOT NULL,
			code TEXT NOT NULL UNIQUE,
			param_type TEXT NOT NULL,
			op TEXT NOT NULL,
			threshold REAL NOT NULL,
			duration_sec INTEGER NOT NULL,
			level TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS compliance_compliance_alerts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rule_id INTEGER NOT NULL,
			tank_id INTEGER NOT NULL,
			sampling_sampling_point_id INTEGER NOT NULL,
			level TEXT NOT NULL,
			message TEXT,
			status TEXT NOT NULL DEFAULT 'open',
			opened_at DATETIME NOT NULL,
			ack_at DATETIME,
			resolved_at DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_compliance_compliance_alerts_tank_status ON compliance_compliance_alerts(tank_id, status)`,
		`CREATE TABLE IF NOT EXISTS calibrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			sensor_id INTEGER NOT NULL,
			performed_at DATETIME NOT NULL,
			due_at DATETIME NOT NULL,
			standard TEXT NOT NULL,
			result TEXT NOT NULL,
			operator TEXT NOT NULL,
			created_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS treatment_cycles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tank_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			product TEXT NOT NULL,
			phase TEXT NOT NULL,
			start_at DATETIME NOT NULL,
			end_at DATETIME,
			status TEXT NOT NULL,
			created_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS compliance_status (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tank_id INTEGER NOT NULL,
			cycle_id INTEGER NOT NULL DEFAULT 0,
			state TEXT NOT NULL,
			reason TEXT,
			changed_at DATETIME NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_compliance_status_tank ON compliance_status(tank_id, changed_at)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

// WithTx 事务执行。
func (db *DB) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}