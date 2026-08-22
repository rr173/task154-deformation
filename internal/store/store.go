package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"deformation/internal/model"
	_ "modernc.org/sqlite"
	"time"
)

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err = s.init(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}
func (s *Store) Close() error { return s.db.Close() }
func (s *Store) init(ctx context.Context) error {
	for _, q := range []string{
		`PRAGMA foreign_keys=ON`,
		`CREATE TABLE IF NOT EXISTS networks(id TEXT PRIMARY KEY, body BLOB NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS points(id TEXT PRIMARY KEY, network_id TEXT NOT NULL, body BLOB NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS periods(id TEXT PRIMARY KEY, network_id TEXT NOT NULL, body BLOB NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS observations(id TEXT PRIMARY KEY, period_id TEXT NOT NULL, fingerprint TEXT NOT NULL, body BLOB NOT NULL, UNIQUE(period_id,fingerprint))`,
		`CREATE TABLE IF NOT EXISTS results(id TEXT PRIMARY KEY, period_id TEXT NOT NULL, version INTEGER NOT NULL, body BLOB NOT NULL, UNIQUE(period_id,version))`,
		`CREATE INDEX IF NOT EXISTS points_network_idx ON points(network_id)`,
		`CREATE INDEX IF NOT EXISTS periods_network_idx ON periods(network_id)`,
		`CREATE INDEX IF NOT EXISTS observations_period_idx ON observations(period_id)`,
		`CREATE INDEX IF NOT EXISTS results_period_idx ON results(period_id, version)`,
	} {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("schema: %w", err)
		}
	}
	return nil
}
func encode(v any) []byte          { b, _ := json.Marshal(v); return b }
func decode(b []byte, v any) error { return json.Unmarshal(b, v) }
func (s *Store) SaveNetwork(ctx context.Context, v model.Network) error {
	_, e := s.db.ExecContext(ctx, `INSERT INTO networks(id,body) VALUES(?,?) ON CONFLICT(id) DO UPDATE SET body=excluded.body`, v.ID, encode(v))
	return e
}
func (s *Store) SavePoint(ctx context.Context, v model.Point) error {
	_, e := s.db.ExecContext(ctx, `INSERT INTO points(id,network_id,body) VALUES(?,?,?) ON CONFLICT(id) DO UPDATE SET network_id=excluded.network_id,body=excluded.body`, v.ID, v.NetworkID, encode(v))
	return e
}
func (s *Store) SavePeriod(ctx context.Context, v model.Period) error {
	_, e := s.db.ExecContext(ctx, `INSERT INTO periods(id,network_id,body) VALUES(?,?,?)`, v.ID, v.NetworkID, encode(v))
	return e
}
func (s *Store) UpdatePeriod(ctx context.Context, v model.Period) error {
	_, e := s.db.ExecContext(ctx, `UPDATE periods SET body=? WHERE id=?`, encode(v), v.ID)
	return e
}
func (s *Store) SaveObservation(ctx context.Context, v model.Observation, fingerprint string) error {
	_, e := s.db.ExecContext(ctx, `INSERT INTO observations(id,period_id,fingerprint,body) VALUES(?,?,?,?)`, v.ID, v.PeriodID, fingerprint, encode(v))
	return e
}
func (s *Store) SaveResult(ctx context.Context, v model.Result) error {
	_, e := s.db.ExecContext(ctx, `INSERT INTO results(id,period_id,version,body) VALUES(?,?,?,?) ON CONFLICT(id) DO UPDATE SET body=excluded.body`, v.ID, v.PeriodID, v.Version, encode(v))
	return e
}
func (s *Store) Network(ctx context.Context, id string) (model.Network, error) {
	var b []byte
	e := s.db.QueryRowContext(ctx, `SELECT body FROM networks WHERE id=?`, id).Scan(&b)
	var v model.Network
	if e == nil {
		e = decode(b, &v)
	}
	return v, e
}
func (s *Store) Period(ctx context.Context, id string) (model.Period, error) {
	var b []byte
	e := s.db.QueryRowContext(ctx, `SELECT body FROM periods WHERE id=?`, id).Scan(&b)
	var v model.Period
	if e == nil {
		e = decode(b, &v)
	}
	return v, e
}
func (s *Store) Points(ctx context.Context, network string) ([]model.Point, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT body FROM points WHERE network_id=? ORDER BY id`, network)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []model.Point{}
	for rows.Next() {
		var b []byte
		var v model.Point
		if e = rows.Scan(&b); e != nil {
			return nil, e
		}
		if e = decode(b, &v); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) Observations(ctx context.Context, period string) ([]model.Observation, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT body FROM observations WHERE period_id=? ORDER BY id`, period)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []model.Observation{}
	for rows.Next() {
		var b []byte
		var v model.Observation
		if e = rows.Scan(&b); e != nil {
			return nil, e
		}
		if e = decode(b, &v); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) Results(ctx context.Context, period string) ([]model.Result, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT body FROM results WHERE period_id=? ORDER BY version`, period)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []model.Result{}
	for rows.Next() {
		var b []byte
		var v model.Result
		if e = rows.Scan(&b); e != nil {
			return nil, e
		}
		if e = decode(b, &v); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
func (s *Store) Now() time.Time                 { return time.Now().UTC() }
