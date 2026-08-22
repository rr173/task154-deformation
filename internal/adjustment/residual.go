package adjustment

import (
	"deformation/internal/model"
	"math"
)

func Residual(estimatedX float64, fixed map[string]model.Point, observation model.Observation) float64 {
	if anchor, ok := fixed[observation.FromPoint]; ok {
		return estimatedX - anchor.X - observation.Distance
	}
	if anchor, ok := fixed[observation.ToPoint]; ok {
		return anchor.X - estimatedX - observation.Distance
	}
	return 0
}

func ClassifyResiduals(points []model.PointResult, observations []model.Observation, threshold float64) ([]model.ObservationIssue, float64) {
	byPoint := make(map[string]model.PointResult, len(points))
	for _, point := range points {
		byPoint[point.PointID] = point
	}
	issues := make([]model.ObservationIssue, 0)
	max := 0.0
	for _, observation := range observations {
		from, hasFrom := byPoint[observation.FromPoint]
		to, hasTo := byPoint[observation.ToPoint]
		if !hasFrom && !hasTo {
			continue
		}
		var magnitude float64
		if hasFrom && hasTo {
			magnitude = math.Abs((to.X - from.X) - observation.Distance)
		} else if hasFrom {
			magnitude = math.Abs(from.Residual)
		} else {
			magnitude = math.Abs(to.Residual)
		}
		if magnitude > max {
			max = magnitude
		}
		if magnitude >= threshold {
			issues = append(issues, model.ObservationIssue{ObservationID: observation.ID, Reason: "residual exceeds precision band", Magnitude: magnitude})
		}
	}
	return issues, max
}
