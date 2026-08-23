package adjustment

import (
	"deformation/internal/model"
	"math"
	"testing"
)

func TestSolve(t *testing.T) {
	out, err := Solve([]model.Point{{ID: "f", Role: model.PointFixed, X: 0}, {ID: "p", Role: model.PointEstimated}}, []model.Observation{{FromPoint: "f", ToPoint: "p", Distance: 10, Precision: 1, Status: model.ObservationValid}})
	if err != nil || len(out) != 1 || out[0].X != 10 {
		t.Fatal(out, err)
	}
}

func TestValidateRejectsNonFiniteFieldImportDistance(t *testing.T) {
	points := []model.Point{{ID: "f", Role: model.PointFixed, X: 0}, {ID: "p", Role: model.PointEstimated}}
	observations := []model.Observation{{ID: "bad", FromPoint: "f", ToPoint: "p", Distance: math.Inf(1), Precision: 1, Status: model.ObservationValid, Source: "field-import"}}
	if err := Validate(points, observations); err == nil {
		t.Fatal("non-finite field-import distance was accepted")
	}
}
