package store

import (
	"context"
	"database/sql"
	"fmt"
	"deformation/internal/model"
)

func (s *Store) Networks(ctx context.Context) ([]model.Network, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT body FROM networks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJSONRows[model.Network](rows)
}

func (s *Store) Periods(ctx context.Context, networkID string) ([]model.Period, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT body FROM periods WHERE network_id=? ORDER BY id`, networkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanJSONRows[model.Period](rows)
}

func (s *Store) Point(ctx context.Context, id string) (model.Point, error) {
	return scanOne[model.Point](ctx, s.db, `SELECT body FROM points WHERE id=?`, id)
}

func (s *Store) Observation(ctx context.Context, id string) (model.Observation, error) {
	return scanOne[model.Observation](ctx, s.db, `SELECT body FROM observations WHERE id=?`, id)
}

func (s *Store) LatestResult(ctx context.Context, periodID string) (model.Result, error) {
	return scanOne[model.Result](ctx, s.db, `SELECT body FROM results WHERE period_id=? ORDER BY version DESC LIMIT 1`, periodID)
}

func (s *Store) UpdateObservation(ctx context.Context, observation model.Observation) error {
	result, err := s.db.ExecContext(ctx, `UPDATE observations SET body=? WHERE id=?`, encode(observation), observation.ID)
	if err != nil {
		return err
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) DeleteObservation(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM observations WHERE id=?`, id)
	if err != nil {
		return err
	}
	if count, _ := result.RowsAffected(); count != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func scanOne[T any](ctx context.Context, db *sql.DB, query string, args ...any) (T, error) {
	var body []byte
	var value T
	err := db.QueryRowContext(ctx, query, args...).Scan(&body)
	if err != nil {
		return value, err
	}
	if err := decode(body, &value); err != nil {
		return value, fmt.Errorf("decode stored record: %w", err)
	}
	return value, nil
}

func scanJSONRows[T any](rows *sql.Rows) ([]T, error) {
	values := make([]T, 0)
	for rows.Next() {
		var body []byte
		if err := rows.Scan(&body); err != nil {
			return nil, err
		}
		var value T
		if err := decode(body, &value); err != nil {
			return nil, fmt.Errorf("decode stored record: %w", err)
		}
		values = append(values, value)
	}
	return values, rows.Err()
}
