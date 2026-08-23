package adjustment

import (
	"deformation/internal/model"
	"reflect"
	"sort"
	"testing"
)

func TestMarkDeformationLeavesInputSnapshotUnchanged(t *testing.T) {
	input := []model.PointResult{
		{PointID: "z", X: 1, Y: 0, Z: 0},
		{PointID: "a", X: 0, Y: 0, Z: 0},
	}
	baseline := []model.PointResult{{PointID: "z", X: 0, Y: 0, Z: 0}, {PointID: "a", X: 0, Y: 0, Z: 0}}
	snapshot := append([]model.PointResult(nil), input...)

	out := MarkDeformation(input, baseline, 0.01)

	if !reflect.DeepEqual(input, snapshot) {
		t.Fatalf("MarkDeformation mutated its input snapshot:\n got = %v\n want = %v", input, snapshot)
	}
	if !out[0].Deformed {
		t.Fatalf("expected the moved point to be flagged deformed, got %v", out)
	}
}

func TestSortResultsLeavesInputSnapshotUnchanged(t *testing.T) {
	input := []model.PointResult{{PointID: "z"}, {PointID: "a"}, {PointID: "m"}}
	snapshot := append([]model.PointResult(nil), input...)

	out := SortResults(input)

	if !reflect.DeepEqual(input, snapshot) {
		t.Fatalf("SortResults mutated its input snapshot:\n got = %v\n want = %v", input, snapshot)
	}
	if !sort.SliceIsSorted(out, func(i, j int) bool { return out[i].PointID < out[j].PointID }) {
		t.Fatalf("SortResults did not return a sorted slice: %v", out)
	}
	// The returned slice must not alias the caller's backing array.
	out[0].PointID = "mutated"
	if input[0].PointID == "mutated" || input[1].PointID == "mutated" {
		t.Fatalf("SortResults returned a slice that aliases the input backing array")
	}
}
