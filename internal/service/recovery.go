package service

import (
	"context"
	"github.com/rr173/task154-deformation/internal/adjustment"
	"github.com/rr173/task154-deformation/internal/model"
)

func (v *Service) Recover(ctx context.Context) (model.RecoveryReport, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	periods, err := v.s.RecoverablePeriods(ctx)
	if err != nil {
		return model.RecoveryReport{}, err
	}
	report := model.RecoveryReport{CheckedAt: v.now().UTC()}
	for _, period := range periods {
		points, pointsErr := v.s.Points(ctx, period.NetworkID)
		observations, observationsErr := v.s.Observations(ctx, period.ID)
		if pointsErr != nil || observationsErr != nil {
			report.FailedPeriodIDs = append(report.FailedPeriodIDs, period.ID)
			continue
		}
		if _, solveErr := adjustment.Solve(points, observations); solveErr != nil {
			period.Status = model.PeriodFailed
			report.FailedPeriodIDs = append(report.FailedPeriodIDs, period.ID)
		} else {
			period.Status = model.PeriodPending
			report.RecoveredPeriodIDs = append(report.RecoveredPeriodIDs, period.ID)
		}
		period.Revision++
		if err := v.s.UpdatePeriod(ctx, period); err != nil {
			return report, err
		}
	}
	return report, nil
}

func (v *Service) previousPublished(ctx context.Context, networkID, excludePeriodID string) ([]model.PointResult, error) {
	periods, err := v.s.Periods(ctx, networkID)
	if err != nil {
		return nil, err
	}
	var selected *model.Period
	for index := range periods {
		period := periods[index]
		if period.ID == excludePeriodID || period.Status != model.PeriodPublished {
			continue
		}
		if selected == nil || period.ObservedAt.After(selected.ObservedAt) {
			copy := period
			selected = &copy
		}
	}
	if selected == nil {
		return nil, nil
	}
	result, err := v.s.LatestResult(ctx, selected.ID)
	if err != nil {
		return nil, err
	}
	return result.Points, nil
}

func meanPrecision(observations []model.Observation) float64 {
	var total float64
	var count int
	for _, observation := range observations {
		if observation.Status.IsIncluded() {
			total += observation.Precision
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}
