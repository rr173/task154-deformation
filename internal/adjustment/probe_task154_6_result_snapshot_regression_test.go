package adjustment

import (
	"deformation/internal/model"
	"testing"
)

func TestBug06_ResultHelpersDoNotMutateCallerSlice(t *testing.T) {
	original := []model.PointResult{{PointID: "z"}, {PointID: "a"}}
	_ = SortResults(original)
	if original[0].PointID != "z" { t.Fatal("sorting mutated caller slice") }
	baseline := []model.PointResult{{PointID: "z", X: 0}, {PointID: "a", X: 0}}
	_ = MarkDeformation(original, baseline, 1)
	if original[0].Deformed { t.Fatal("deformation marking mutated caller slice") }
}
