package service

import (
	"context"
	"deformation/internal/model"
	"deformation/internal/store"
	"path/filepath"
	"testing"
	"time"
)

func TestBug07_CannotPublishAfterOnlyObservationIsWithdrawn(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "withdrawn.db"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	service := New(db)
	ctx := context.Background()
	network, err := service.CreateNetwork(ctx, "site")
	if err != nil { t.Fatal(err) }
	for _, point := range []model.Point{{ID: "fixed", NetworkID: network.ID, Role: model.PointFixed}, {ID: "target", NetworkID: network.ID, Role: model.PointEstimated}} {
		if _, err = service.AddPoint(ctx, point); err != nil { t.Fatal(err) }
	}
	period, err := service.CreatePeriod(ctx, network.ID, time.Now())
	if err != nil { t.Fatal(err) }
	observation, err := service.Import(ctx, model.Observation{PeriodID: period.ID, FromPoint: "fixed", ToPoint: "target", Distance: 10, Precision: 0.1})
	if err != nil { t.Fatal(err) }
	if _, err = service.Compute(ctx, period.ID); err != nil { t.Fatal(err) }
	if _, err = service.WithdrawObservation(ctx, observation.ID); err != nil { t.Fatal(err) }
	if _, err = service.Publish(ctx, period.ID); err == nil { t.Fatal("period with no usable observations was published") }
}
