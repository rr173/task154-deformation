package adjustment

import (
	"deformation/internal/model"
	"testing"
)

func TestSolve(t *testing.T) {
	out, err := Solve([]model.Point{{ID: "f", Role: model.PointFixed, X: 0}, {ID: "p", Role: model.PointEstimated}}, []model.Observation{{FromPoint: "f", ToPoint: "p", Distance: 10, Precision: 1, Status: model.ObservationValid}})
	if err != nil || len(out) != 1 || out[0].X != 10 {
		t.Fatal(out, err)
	}
}
