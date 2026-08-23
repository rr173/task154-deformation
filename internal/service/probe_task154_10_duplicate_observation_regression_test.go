package service

import (
	"context"
	"deformation/internal/model"
	"deformation/internal/store"
	"path/filepath"
	"testing"
)

func TestBug10_RepeatedObservationIsReportedAsDuplicate(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "duplicate.db"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	service := New(db)
	ctx := context.Background()
	network, err := service.CreateNetwork(ctx, "site")
	if err != nil { t.Fatal(err) }
	for _, point := range []model.Point{{ID: "fixed", NetworkID: network.ID, Role: model.PointFixed}, {ID: "target", NetworkID: network.ID, Role: model.PointEstimated}} {
		if _, err = service.AddPoint(ctx, point); err != nil { t.Fatal(err) }
	}
	period, err := service.CreatePeriod(ctx, network.ID, service.now())
	if err != nil { t.Fatal(err) }
	input := model.Observation{PeriodID: period.ID, FromPoint: "fixed", ToPoint: "target", Distance: 10, Precision: 0.1}
	if _, err = service.Import(ctx, input); err != nil { t.Fatal(err) }
	second, err := service.Import(ctx, input)
	if err != nil { t.Fatal(err) }
	if second.Status != model.ObservationDuplicate { t.Fatalf("second import status=%s, want duplicate", second.Status) }
}
