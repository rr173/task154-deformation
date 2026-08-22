package service

import (
	"context"
	"fmt"
	"github.com/rr173/task154-deformation/internal/model"
)

func (v *Service) WithdrawObservation(ctx context.Context, id string) (model.Observation, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	observation, err := v.s.Observation(ctx, id)
	if err != nil {
		return model.Observation{}, err
	}
	period, err := v.s.Period(ctx, observation.PeriodID)
	if err != nil {
		return model.Observation{}, err
	}
	if period.Status.IsTerminal() {
		return model.Observation{}, fmt.Errorf("published period is immutable")
	}
	if observation.Status == model.ObservationWithdrawn {
		return observation, nil
	}
	observation.Status = model.ObservationWithdrawn
	if err := v.s.UpdateObservation(ctx, observation); err != nil {
		return model.Observation{}, err
	}
	if period.Status == model.PeriodSucceeded {
		if err := model.ValidateTransition(period.Status, model.PeriodPending); err != nil {
			return model.Observation{}, err
		}
		period.Status = model.PeriodPending
		period.Revision++
		if err := v.s.UpdatePeriod(ctx, period); err != nil {
			return model.Observation{}, err
		}
	}
	return observation, nil
}

func (v *Service) RestoreObservation(ctx context.Context, id string) (model.Observation, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	observation, err := v.s.Observation(ctx, id)
	if err != nil {
		return model.Observation{}, err
	}
	period, err := v.s.Period(ctx, observation.PeriodID)
	if err != nil {
		return model.Observation{}, err
	}
	if period.Status.IsTerminal() {
		return model.Observation{}, fmt.Errorf("published period is immutable")
	}
	if observation.Status != model.ObservationWithdrawn {
		return observation, nil
	}
	observation.Status = model.ObservationValid
	if err := v.s.UpdateObservation(ctx, observation); err != nil {
		return model.Observation{}, err
	}
	if period.Status != model.PeriodPending {
		if err := model.ValidateTransition(period.Status, model.PeriodPending); err != nil {
			return model.Observation{}, err
		}
		period.Status = model.PeriodPending
		period.Revision++
		if err := v.s.UpdatePeriod(ctx, period); err != nil {
			return model.Observation{}, err
		}
	}
	return observation, nil
}

func (v *Service) Observations(ctx context.Context, periodID string) ([]model.Observation, error) {
	return v.s.Observations(ctx, periodID)
}

func (v *Service) SetNetworkReady(ctx context.Context, networkID string) (model.Network, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	network, err := v.s.Network(ctx, networkID)
	if err != nil {
		return model.Network{}, err
	}
	if !network.Status.CanAddData() {
		return model.Network{}, fmt.Errorf("network cannot be marked ready from status %s", network.Status)
	}
	points, err := v.s.Points(ctx, networkID)
	if err != nil {
		return model.Network{}, err
	}
	fixed := 0
	for _, point := range points {
		if point.Role == model.PointFixed {
			fixed++
		}
	}
	if fixed == 0 {
		return model.Network{}, fmt.Errorf("network needs a fixed point before it is ready")
	}
	network.Status = model.NetworkReady
	return network, v.s.SaveNetwork(ctx, network)
}

func (v *Service) ArchiveNetwork(ctx context.Context, networkID string) (model.Network, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	network, err := v.s.Network(ctx, networkID)
	if err != nil {
		return model.Network{}, err
	}
	if !network.Status.CanArchive() {
		return model.Network{}, fmt.Errorf("network cannot be archived from status %s", network.Status)
	}
	periods, err := v.s.Periods(ctx, networkID)
	if err != nil {
		return model.Network{}, err
	}
	for _, period := range periods {
		if period.Status == model.PeriodRunning {
			return model.Network{}, fmt.Errorf("network has a running observation period")
		}
	}
	network.Status = model.NetworkArchived
	if err := v.s.SaveNetwork(ctx, network); err != nil {
		return model.Network{}, err
	}
	return network, nil
}
