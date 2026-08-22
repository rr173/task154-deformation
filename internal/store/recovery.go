package store

import (
	"context"
	"deformation/internal/model"
)

func (s *Store) RecoverablePeriods(ctx context.Context) ([]model.Period, error) {
	networks, err := s.Networks(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]model.Period, 0)
	for _, network := range networks {
		periods, err := s.Periods(ctx, network.ID)
		if err != nil {
			return nil, err
		}
		for _, period := range periods {
			if period.Status == model.PeriodRunning {
				result = append(result, period)
			}
		}
	}
	return result, nil
}

func (s *Store) NetworkExists(ctx context.Context, id string) (bool, error) {
	_, err := s.Network(ctx, id)
	if err == nil {
		return true, nil
	}
	if isNoRows(err) {
		return false, nil
	}
	return false, err
}

func isNoRows(err error) bool {
	return err != nil && err.Error() == "sql: no rows in result set"
}
