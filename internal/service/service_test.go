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

// ArchiveNetwork must be an irreversible closure of data ingestion: the
// network must never return to a ready state, re-archive, or accept fresh
// points, observation periods, or observations.
func TestArchiveNetworkIsIrreversible(t *testing.T) {
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
	if _, err := service.AddPoint(ctx, model.Point{ID: "f", NetworkID: network.ID, Role: model.PointFixed, X: 0}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetNetworkReady(ctx, network.ID); err != nil {
		t.Fatal(err)
	}
	// Mark ready, open a period, then archive. Archiving from ready must succeed.
	period, err := service.CreatePeriod(ctx, network.ID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Import(ctx, model.Observation{PeriodID: period.ID, FromPoint: "f", ToPoint: "f", Distance: 1, Precision: 0.1}); err == nil {
		t.Fatal("self-observation should be rejected before archiving")
	}
	archived, err := service.ArchiveNetwork(ctx, network.ID)
	if err != nil {
		t.Fatal(err)
	}
	if archived.Status != model.NetworkArchived {
		t.Fatalf("network status = %s, want archived", archived.Status)
	}

	// Re-marking ready must be refused — archive is terminal.
	if _, err := service.SetNetworkReady(ctx, network.ID); err == nil {
		t.Fatal("archived network was marked ready again")
	}
	// Re-archiving an already-archived network must also be refused.
	if _, err := service.ArchiveNetwork(ctx, network.ID); err == nil {
		t.Fatal("archived network was archived again")
	}
	// Fresh points, periods, and observations must all be refused.
	if _, err := service.AddPoint(ctx, model.Point{ID: "p", NetworkID: network.ID, Role: model.PointEstimated}); err == nil {
		t.Fatal("archived network accepted a new point")
	}
	if _, err := service.CreatePeriod(ctx, network.ID, time.Now()); err == nil {
		t.Fatal("archived network accepted a new period")
	}
	if _, err := service.Import(ctx, model.Observation{PeriodID: period.ID, FromPoint: "f", ToPoint: "p", Distance: 12, Precision: 0.1}); err == nil {
		t.Fatal("archived network accepted a new observation")
	}

	// The persisted status must remain archived — no path revived it.
	persisted, err := service.Network(ctx, network.ID)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.Status != model.NetworkArchived {
		t.Fatalf("persisted network status = %s, want archived", persisted.Status)
	}
}
