package service

import (
	"fmt"
	"deformation/internal/model"
	"math"
)

const coordinateLimit = 100_000_000.0

func validatePointCoordinates(point model.Point) error {
	coordinates := []struct {
		name  string
		value float64
	}{
		{name: "x", value: point.X},
		{name: "y", value: point.Y},
		{name: "z", value: point.Z},
	}
	for _, coordinate := range coordinates {
		if math.IsNaN(coordinate.value) || math.IsInf(coordinate.value, 0) {
			return fmt.Errorf("point %s coordinate must be finite", coordinate.name)
		}
		if math.Abs(coordinate.value) > coordinateLimit {
			return fmt.Errorf("point %s coordinate exceeds supported range", coordinate.name)
		}
	}
	return nil
}

func validateObservationForPeriod(period model.Period, from, to model.Point, observation model.Observation) error {
	if from.ID == to.ID {
		return fmt.Errorf("observation must connect two distinct points")
	}
	if from.NetworkID != period.NetworkID || to.NetworkID != period.NetworkID {
		return fmt.Errorf("observation points must belong to the period network")
	}
	if from.Role == model.PointDisabled || to.Role == model.PointDisabled {
		return fmt.Errorf("observation cannot reference a disabled point")
	}
	if math.IsNaN(observation.Distance) || math.IsInf(observation.Distance, 0) || math.IsNaN(observation.Precision) || math.IsInf(observation.Precision, 0) {
		return fmt.Errorf("observation distance and precision must be finite")
	}
	if observation.Distance <= 0 || observation.Precision <= 0 {
		return fmt.Errorf("distance and precision must be positive")
	}
	return nil
}

func canCreatePeriod(network model.Network) error {
	if !network.Status.CanAddData() {
		return fmt.Errorf("network does not accept new observation periods")
	}
	return nil
}
