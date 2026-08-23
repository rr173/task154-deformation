package service

import (
	"context"
	"deformation/internal/model"
	"deformation/internal/store"
	"path/filepath"
	"testing"
	"time"
)

func TestComputeWithdrawAndPublish(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "service.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	service := New(db)
	ctx := context.Background()
	network, err := service.CreateNetwork(ctx, "site")
	if err != nil {
		t.Fatal(err)
	}
	for _, point := range []model.Point{{ID: "f", NetworkID: network.ID, Role: model.PointFixed}, {ID: "p", NetworkID: network.ID, Role: model.PointEstimated}} {
		if _, err := service.AddPoint(ctx, point); err != nil {
			t.Fatal(err)
		}
	}
	period, err := service.CreatePeriod(ctx, network.ID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	observation, err := service.Import(ctx, model.Observation{PeriodID: period.ID, FromPoint: "f", ToPoint: "p", Distance: 12, Precision: 0.1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Compute(ctx, period.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.WithdrawObservation(ctx, observation.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RestoreObservation(ctx, observation.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Compute(ctx, period.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Publish(ctx, period.ID); err != nil {
		t.Fatal(err)
	}

	published, err := service.s.Period(ctx, period.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !published.Status.IsTerminal() {
		t.Fatalf("published period status %s is not terminal", published.Status)
	}
	if published.Status.CanAcceptObservation() {
		t.Fatal("published period still accepts observations")
	}
	if _, err := service.Import(ctx, model.Observation{PeriodID: period.ID, FromPoint: "f", ToPoint: "p", Distance: 99, Precision: 0.1, Source: "field-import"}); err == nil {
		t.Fatal("published period accepted a new observation")
	}
	if _, err := service.WithdrawObservation(ctx, observation.ID); err == nil {
		t.Fatal("published period allowed withdrawing an existing observation")
	}
	if _, err := service.RestoreObservation(ctx, observation.ID); err == nil {
		t.Fatal("published period allowed restoring an existing observation")
	}
}

func TestImportRejectsPointFromAnotherNetwork(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "service.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	service := New(db)
	ctx := context.Background()
	first, err := service.CreateNetwork(ctx, "first")
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.CreateNetwork(ctx, "second")
	if err != nil {
		t.Fatal(err)
	}
	for _, point := range []model.Point{
		{ID: "first-fixed", NetworkID: first.ID, Role: model.PointFixed},
		{ID: "second-estimated", NetworkID: second.ID, Role: model.PointEstimated},
	} {
		if _, err := service.AddPoint(ctx, point); err != nil {
			t.Fatal(err)
		}
	}
	period, err := service.CreatePeriod(ctx, first.ID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Import(ctx, model.Observation{PeriodID: period.ID, FromPoint: "first-fixed", ToPoint: "second-estimated", Distance: 12, Precision: 0.1}); err == nil {
		t.Fatal("expected cross-network observation to be rejected")
	}
}

func TestArchiveNetworkBlocksNewPointsAndPeriods(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "service.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	service := New(db)
	ctx := context.Background()
	network, err := service.CreateNetwork(ctx, "site")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ArchiveNetwork(ctx, network.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddPoint(ctx, model.Point{NetworkID: network.ID, Role: model.PointFixed}); err == nil {
		t.Fatal("archived network accepted a new point")
	}
	if _, err := service.CreatePeriod(ctx, network.ID, time.Now()); err == nil {
		t.Fatal("archived network accepted a new period")
	}
}
