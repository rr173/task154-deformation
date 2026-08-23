package adjustment

import (
	"fmt"
	"deformation/internal/model"
	"math"
)

func Validate(points []model.Point, observations []model.Observation) error {
	known := make(map[string]model.Point, len(points))
	fixed := 0
	for _, point := range points {
		if point.ID == "" {
			return fmt.Errorf("point id is empty")
		}
		if _, exists := known[point.ID]; exists {
			return fmt.Errorf("point %s appears twice", point.ID)
		}
		if point.Role == model.PointFixed {
			fixed++
		}
		known[point.ID] = point
	}
	if fixed == 0 {
		return fmt.Errorf("at least one fixed point is required")
	}
	if len(observations) == 0 {
		return fmt.Errorf("no observations imported")
	}
	valid := 0
	for _, observation := range observations {
		if !observation.Status.IsIncluded() {
			continue
		}
		valid++
		if _, ok := known[observation.FromPoint]; !ok {
			return fmt.Errorf("observation %s references unknown source point", observation.ID)
		}
		if _, ok := known[observation.ToPoint]; !ok {
			return fmt.Errorf("observation %s references unknown target point", observation.ID)
		}
		if observation.FromPoint == observation.ToPoint {
			return fmt.Errorf("observation %s has identical endpoints", observation.ID)
		}
		if math.IsNaN(observation.Distance) || math.IsInf(observation.Distance, 0) {
			return fmt.Errorf("observation %s has non-finite distance", observation.ID)
		}
		if observation.Distance <= 0 || observation.Precision <= 0 {
			return fmt.Errorf("observation %s has invalid distance or precision", observation.ID)
		}
	}
	if valid == 0 {
		return fmt.Errorf("all observations are withdrawn")
	}
	return nil
}

func ConnectedToFixed(points []model.Point, observations []model.Observation) map[string]bool {
	neighbors := map[string][]string{}
	fixed := map[string]bool{}
	for _, point := range points {
		if point.Role == model.PointFixed {
			fixed[point.ID] = true
		}
	}
	for _, observation := range observations {
		if !observation.Status.IsIncluded() {
			continue
		}
		neighbors[observation.FromPoint] = append(neighbors[observation.FromPoint], observation.ToPoint)
		neighbors[observation.ToPoint] = append(neighbors[observation.ToPoint], observation.FromPoint)
	}
	queue := make([]string, 0, len(fixed))
	for id := range fixed {
		queue = append(queue, id)
	}
	reachable := map[string]bool{}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if reachable[id] {
			continue
		}
		reachable[id] = true
		queue = append(queue, neighbors[id]...)
	}
	return reachable
}
