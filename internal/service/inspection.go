package service

import (
	"context"
	"fmt"
	"deformation/internal/adjustment"
	"deformation/internal/model"
)

func (v *Service) InspectNetwork(ctx context.Context, networkID string) (model.NetworkInspection, error) {
	network, err := v.s.Network(ctx, networkID)
	if err != nil {
		return model.NetworkInspection{}, err
	}
	points, err := v.s.Points(ctx, network.ID)
	if err != nil {
		return model.NetworkInspection{}, err
	}
	periods, err := v.s.Periods(ctx, network.ID)
	if err != nil {
		return model.NetworkInspection{}, err
	}
	inspection := model.NetworkInspection{NetworkID: network.ID, InspectedAt: v.now().UTC(), Ready: network.Status == model.NetworkReady || network.Status == model.NetworkPublished, Periods: make([]model.PeriodInspection, 0, len(periods))}
	allObservations := make([]model.Observation, 0)
	coverage := make(map[string]int, len(points))
	for _, point := range points {
		switch point.Role {
		case model.PointFixed:
			inspection.FixedPointCount++
		case model.PointEstimated:
			inspection.EstimatedPointCount++
		case model.PointDisabled:
			inspection.DisabledPointCount++
		}
	}
	for _, period := range periods {
		row, observations, err := v.inspectPeriod(ctx, period)
		if err != nil {
			return inspection, err
		}
		inspection.Periods = append(inspection.Periods, row)
		allObservations = append(allObservations, observations...)
		for _, observation := range observations {
			if !observation.Status.IsIncluded() {
				continue
			}
			coverage[observation.FromPoint]++
			coverage[observation.ToPoint]++
		}
	}
	inspection.ObservationCount = len(allObservations)
	reachable := adjustment.ConnectedToFixed(points, allObservations)
	for _, point := range points {
		if point.Role == model.PointDisabled {
			continue
		}
		if reachable[point.ID] {
			inspection.ConnectedPointCount++
		} else {
			inspection.MissingPointIDs = append(inspection.MissingPointIDs, point.ID)
		}
		if point.Role == model.PointEstimated && coverage[point.ID] == 0 {
			inspection.MissingPointIDs = appendUnique(inspection.MissingPointIDs, point.ID)
		}
	}
	inspection.ReadinessWarnings = readinessWarnings(network, inspection)
	inspection.Ready = inspection.Ready && inspection.IsHealthy()
	return inspection, nil
}

func readinessWarnings(network model.Network, inspection model.NetworkInspection) []string {
	warnings := make([]string, 0, 4)
	if inspection.FixedPointCount == 0 {
		warnings = append(warnings, "network has no fixed reference point")
	}
	if inspection.EstimatedPointCount == 0 {
		warnings = append(warnings, "network has no estimated point")
	}
	if len(inspection.MissingPointIDs) > 0 {
		warnings = append(warnings, "some active points are not connected to a fixed reference")
	}
	if inspection.ObservationCount == 0 {
		warnings = append(warnings, "network has no observations")
	}
	if network.Status == model.NetworkDraft && len(warnings) == 0 {
		warnings = append(warnings, "network is structurally ready but has not been marked ready")
	}
	return warnings
}

func (v *Service) InspectPointCoverage(ctx context.Context, networkID string) ([]model.PointCoverage, error) {
	points, err := v.s.Points(ctx, networkID)
	if err != nil {
		return nil, err
	}
	periods, err := v.s.Periods(ctx, networkID)
	if err != nil {
		return nil, err
	}
	coverage := make(map[string]model.PointCoverage, len(points))
	for _, point := range points {
		coverage[point.ID] = model.PointCoverage{PointID: point.ID, Role: point.Role}
	}
	allObservations := make([]model.Observation, 0)
	for _, period := range periods {
		observations, err := v.s.Observations(ctx, period.ID)
		if err != nil {
			return nil, err
		}
		for _, observation := range observations {
			if !observation.Status.IsIncluded() {
				continue
			}
			allObservations = append(allObservations, observation)
			for _, pointID := range []string{observation.FromPoint, observation.ToPoint} {
				row := coverage[pointID]
				row.ObservationCount++
				if row.LatestPeriodID == "" || period.ObservedAt.After(periodTime(periods, row.LatestPeriodID).ObservedAt) {
					row.LatestPeriodID = period.ID
				}
				coverage[pointID] = row
			}
		}
	}
	reachable := adjustment.ConnectedToFixed(points, allObservations)
	result := make([]model.PointCoverage, 0, len(points))
	for _, point := range points {
		row := coverage[point.ID]
		row.Connected = reachable[point.ID]
		result = append(result, row)
	}
	return result, nil
}

func (v *Service) inspectPeriod(ctx context.Context, period model.Period) (model.PeriodInspection, []model.Observation, error) {
	observations, err := v.s.Observations(ctx, period.ID)
	if err != nil {
		return model.PeriodInspection{}, nil, err
	}
	row := model.PeriodInspection{PeriodID: period.ID, Status: period.Status, ObservedAt: period.ObservedAt, ObservationCount: len(observations), Warnings: make([]string, 0)}
	for _, observation := range observations {
		if observation.Status.IsIncluded() {
			row.IncludedCount++
		}
		if observation.Status == model.ObservationWithdrawn {
			row.WithdrawnCount++
		}
	}
	result, resultErr := v.s.LatestResult(ctx, period.ID)
	if resultErr == nil {
		row.HasResult = true
		row.Published = result.Published
	}
	if period.Status == model.PeriodPublished && !row.Published {
		row.Warnings = append(row.Warnings, "period is published but its latest result is not marked published")
	}
	if period.Status == model.PeriodSucceeded && !row.HasResult {
		row.Warnings = append(row.Warnings, "successful period has no persisted result")
	}
	if row.IncludedCount == 0 && !period.Status.IsTerminal() {
		row.Warnings = append(row.Warnings, "period has no included observations")
	}
	if row.WithdrawnCount > 0 && period.Status == model.PeriodSucceeded {
		row.Warnings = append(row.Warnings, "withdrawn observations require a recalculation")
	}
	return row, observations, nil
}

func (v *Service) ValidatePublication(ctx context.Context, periodID string) error {
	period, err := v.s.Period(ctx, periodID)
	if err != nil {
		return err
	}
	if !period.Status.CanPublish() {
		return fmt.Errorf("period %s is not ready for publication", periodID)
	}
	observations, err := v.s.Observations(ctx, periodID)
	if err != nil {
		return err
	}
	if len(observations) == 0 {
		return fmt.Errorf("period %s has no usable observations", periodID)
	}
	result, err := v.s.LatestResult(ctx, periodID)
	if err != nil {
		return fmt.Errorf("period %s has no computed result: %w", periodID, err)
	}
	if result.MaxResidual > meanPrecision(observations)*5 {
		return fmt.Errorf("period %s residual exceeds publication tolerance", periodID)
	}
	return nil
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func periodTime(periods []model.Period, id string) model.Period {
	for _, period := range periods {
		if period.ID == id {
			return period
		}
	}
	return model.Period{}
}
