package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"deformation/internal/model"
)

func (v *Service) CreateNetworkInput(ctx context.Context, input model.NetworkInput) (model.Network, error) {
	if err := input.Validate(); err != nil {
		return model.Network{}, err
	}
	network := model.Network{ID: uuid.NewString(), Name: input.Name, Description: input.Description, CoordinateDatum: input.Datum, Status: model.NetworkDraft, CreatedAt: v.now().UTC()}
	if network.CoordinateDatum == "" {
		network.CoordinateDatum = "local-engineering"
	}
	return network, v.s.SaveNetwork(ctx, network)
}

func (v *Service) AddPointInput(ctx context.Context, networkID string, input model.PointInput) (model.Point, error) {
	if err := input.Validate(); err != nil {
		return model.Point{}, err
	}
	point := model.Point{ID: input.ID, NetworkID: networkID, Role: input.Role, X: input.X, Y: input.Y, Z: input.Z, Label: input.Label}
	return v.AddPoint(ctx, point)
}

func (v *Service) CreatePeriodInput(ctx context.Context, networkID string, input model.PeriodInput) (model.Period, error) {
	if len(input.Note) > 1000 {
		return model.Period{}, fmt.Errorf("period note is too long")
	}
	period, err := v.CreatePeriod(ctx, networkID, input.Normalize(v.now()))
	if err != nil {
		return period, err
	}
	period.Note = input.Note
	if err := v.s.UpdatePeriod(ctx, period); err != nil {
		return period, err
	}
	return period, nil
}

func (v *Service) ImportInput(ctx context.Context, periodID string, input model.ObservationInput) (model.Observation, error) {
	if err := input.Validate(); err != nil {
		return model.Observation{}, err
	}
	return v.Import(ctx, model.Observation{ID: input.ID, PeriodID: periodID, FromPoint: input.FromPoint, ToPoint: input.ToPoint, Distance: input.Distance, Precision: input.Precision, Source: input.Source})
}

func (v *Service) Networks(ctx context.Context) ([]model.NetworkSummary, error) {
	networks, err := v.s.Networks(ctx)
	if err != nil {
		return nil, err
	}
	summaries := make([]model.NetworkSummary, 0, len(networks))
	for _, network := range networks {
		points, err := v.s.Points(ctx, network.ID)
		if err != nil {
			return nil, err
		}
		periods, err := v.s.Periods(ctx, network.ID)
		if err != nil {
			return nil, err
		}
		summary := model.NetworkSummary{Network: network, PointCount: len(points), PeriodCount: len(periods)}
		for index := range periods {
			period := periods[index]
			if period.Status == model.PeriodPublished {
				summary.PublishedPeriods++
			}
			if summary.LatestPeriod == nil || period.ObservedAt.After(summary.LatestPeriod.ObservedAt) {
				copy := period
				summary.LatestPeriod = &copy
			}
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func (v *Service) Network(ctx context.Context, id string) (model.Network, error) {
	return v.s.Network(ctx, id)
}
func (v *Service) Points(ctx context.Context, networkID string) ([]model.Point, error) {
	return v.s.Points(ctx, networkID)
}
func (v *Service) Periods(ctx context.Context, networkID string) ([]model.Period, error) {
	return v.s.Periods(ctx, networkID)
}
func (v *Service) Result(ctx context.Context, periodID string) (model.Result, error) {
	return v.s.LatestResult(ctx, periodID)
}

func (v *Service) PeriodSummary(ctx context.Context, periodID string) (model.PeriodSummary, error) {
	period, err := v.s.Period(ctx, periodID)
	if err != nil {
		return model.PeriodSummary{}, err
	}
	observations, err := v.s.Observations(ctx, periodID)
	if err != nil {
		return model.PeriodSummary{}, err
	}
	summary := model.PeriodSummary{Period: period, ObservationCount: len(observations)}
	for _, observation := range observations {
		switch observation.Status {
		case model.ObservationValid:
			summary.ValidObservations++
		case model.ObservationOutlier:
			summary.OutlierObservations++
		case model.ObservationWithdrawn:
			summary.WithdrawnObservations++
		}
	}
	result, err := v.s.LatestResult(ctx, periodID)
	if err == nil {
		summary.LatestResult = &result
	} else if err != sql.ErrNoRows {
		return summary, err
	}
	return summary, nil
}
