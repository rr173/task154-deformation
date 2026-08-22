package service

import (
	"context"
	"github.com/rr173/task154-deformation/internal/model"
)

func (v *Service) markOutliers(ctx context.Context, observations []model.Observation, issues []model.ObservationIssue) error {
	issueByID := make(map[string]model.ObservationIssue, len(issues))
	for _, issue := range issues {
		issueByID[issue.ObservationID] = issue
	}
	for _, observation := range observations {
		issue, flagged := issueByID[observation.ID]
		if !flagged {
			continue
		}
		observation.Residual = issue.Magnitude
		if observation.Status == model.ObservationValid {
			observation.Status = model.ObservationOutlier
		}
		if err := v.s.UpdateObservation(ctx, observation); err != nil {
			return err
		}
	}
	return nil
}

func (v *Service) QualityIssues(ctx context.Context, periodID string) ([]model.ObservationIssue, error) {
	period, err := v.s.Period(ctx, periodID)
	if err != nil {
		return nil, err
	}
	result, err := v.s.LatestResult(ctx, periodID)
	if err != nil {
		return nil, err
	}
	observations, err := v.s.Observations(ctx, period.ID)
	if err != nil {
		return nil, err
	}
	issues, _ := classifyStoredIssues(result, observations)
	return issues, nil
}

func classifyStoredIssues(result model.Result, observations []model.Observation) ([]model.ObservationIssue, float64) {
	issues := make([]model.ObservationIssue, 0)
	max := 0.0
	for _, observation := range observations {
		if observation.Status != model.ObservationOutlier {
			continue
		}
		magnitude := observation.Residual
		if magnitude > max {
			max = magnitude
		}
		issues = append(issues, model.ObservationIssue{ObservationID: observation.ID, Reason: "residual exceeded the precision band during result version " + itoa(result.Version), Magnitude: magnitude})
	}
	return issues, max
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}
