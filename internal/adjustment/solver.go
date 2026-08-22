package adjustment

import (
	"errors"
	"deformation/internal/model"
	"math"
)

func Solve(points []model.Point, observations []model.Observation) ([]model.PointResult, error) {
	if err := Validate(points, observations); err != nil {
		return nil, err
	}
	reachable := ConnectedToFixed(points, observations)
	fixed := map[string]model.Point{}
	estimated := map[string]model.Point{}
	for _, p := range points {
		if p.Role == model.PointFixed {
			fixed[p.ID] = p
		}
		if p.Role == model.PointEstimated {
			estimated[p.ID] = p
		}
	}
	if len(fixed) == 0 {
		return nil, errors.New("at least one fixed point is required")
	}
	if len(observations) == 0 {
		return nil, errors.New("no valid observations")
	}
	out := make([]model.PointResult, 0, len(estimated))
	for id, p := range estimated {
		if !reachable[id] {
			return nil, errors.New("network is disconnected for point " + id)
		}
		sum, n := 0.0, 0
		for _, o := range observations {
			if !o.Status.IsIncluded() {
				continue
			}
			if o.FromPoint == id {
				if anchor, ok := fixed[o.ToPoint]; ok {
					sum += anchor.X - o.Distance
					n++
				}
			}
			if o.ToPoint == id {
				if anchor, ok := fixed[o.FromPoint]; ok {
					sum += anchor.X + o.Distance
					n++
				}
			}
		}
		if n == 0 {
			return nil, errors.New("network is unsolvable for point " + id)
		}
		x := sum / float64(n)
		std, residual := 0.0, 0.0
		for _, o := range observations {
			if o.FromPoint == id || o.ToPoint == id {
				std += o.Precision * o.Precision
				residual += Residual(x, fixed, o)
			}
		}
		std = math.Sqrt(std / float64(n))
		out = append(out, model.PointResult{PointID: id, X: x, Y: p.Y, Z: p.Z, StdDev: std, Residual: residual / float64(n)})
	}
	return out, nil
}
func Compare(before, after []model.PointResult, threshold float64) []model.Compare {
	old := map[string]model.PointResult{}
	for _, p := range before {
		old[p.PointID] = p
	}
	out := []model.Compare{}
	for _, p := range after {
		if b, ok := old[p.PointID]; ok {
			dx, dy, dz := p.X-b.X, p.Y-b.Y, p.Z-b.Z
			d := math.Sqrt(dx*dx + dy*dy + dz*dz)
			out = append(out, model.Compare{PointID: p.PointID, DX: dx, DY: dy, DZ: dz, Distance: d, Significant: d >= threshold})
		}
	}
	return out
}
