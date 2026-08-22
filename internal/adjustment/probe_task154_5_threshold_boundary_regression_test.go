package adjustment

import (
	"deformation/internal/model"
	"testing"
)

func TestBug05_ThresholdEqualityIsNotAnAlarm(t *testing.T) {
	before := []model.PointResult{{PointID: "p", X: 0}}
	after := []model.PointResult{{PointID: "p", X: 1}}
	if Compare(before, after, 1)[0].Significant { t.Error("equal displacement was marked significant") }
	if MarkDeformation(after, before, 1)[0].Deformed { t.Error("equal displacement was marked deformed") }
	points := []model.PointResult{{PointID: "fixed", X: 0}, {PointID: "p", X: 1}}
	observations := []model.Observation{{ID: "o", FromPoint: "fixed", ToPoint: "p", Distance: 0, Status: model.ObservationValid}}
	issues, _ := ClassifyResiduals(points, observations, 1)
	if len(issues) != 0 { t.Fatal("equal residual was marked as an issue") }
}
