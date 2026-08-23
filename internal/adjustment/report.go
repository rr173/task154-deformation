package adjustment

import (
	"deformation/internal/model"
	"math"
	"sort"
)

func SortResults(results []model.PointResult) []model.PointResult {
	out := make([]model.PointResult, len(results))
	copy(out, results)
	sort.Slice(out, func(i, j int) bool { return out[i].PointID < out[j].PointID })
	return out
}

func MarkDeformation(results []model.PointResult, baseline []model.PointResult, threshold float64) []model.PointResult {
	old := make(map[string]model.PointResult, len(baseline))
	for _, point := range baseline {
		old[point.PointID] = point
	}
	out := make([]model.PointResult, len(results))
	copy(out, results)
	for index := range out {
		if prior, ok := old[out[index].PointID]; ok {
			dx, dy, dz := out[index].X-prior.X, out[index].Y-prior.Y, out[index].Z-prior.Z
			out[index].Deformed = math.Sqrt(dx*dx+dy*dy+dz*dz) > threshold
		}
	}
	return out
}

func ObservationCount(observations []model.Observation) int {
	count := 0
	for _, observation := range observations {
		if observation.Status.IsIncluded() {
			count++
		}
	}
	return count
}
