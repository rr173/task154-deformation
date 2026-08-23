package service

import (
	"context"
	"deformation/internal/model"
	"deformation/internal/store"
	"path/filepath"
	"testing"
	"time"
)

// TestPublishBlockedAfterWithdrawingOnlyObservation reproduces the reported defect:
// after the sole usable observation of a period is withdrawn, publishing must be
// refused rather than reusing the stale, previously computed result.
func TestPublishBlockedAfterWithdrawingOnlyObservation(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "publish-guard.db"))
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
	for _, point := range []model.Point{
		{ID: "f", NetworkID: network.ID, Role: model.PointFixed},
		{ID: "p", NetworkID: network.ID, Role: model.PointEstimated},
	} {
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
	if _, err := service.Publish(ctx, period.ID); err == nil {
		t.Fatal("expected publish to be refused after the only observation was withdrawn, but it succeeded")
	}
}

// TestPublishBlockedAfterAllObservationsWithdrawn covers the multi-observation case:
// withdrawing every usable observation must also block publication.
func TestPublishBlockedAfterAllObservationsWithdrawn(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "publish-guard-multi.db"))
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
	for _, point := range []model.Point{
		{ID: "f", NetworkID: network.ID, Role: model.PointFixed},
		{ID: "p", NetworkID: network.ID, Role: model.PointEstimated},
	} {
		if _, err := service.AddPoint(ctx, point); err != nil {
			t.Fatal(err)
		}
	}
	period, err := service.CreatePeriod(ctx, network.ID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	first, err := service.Import(ctx, model.Observation{PeriodID: period.ID, FromPoint: "f", ToPoint: "p", Distance: 12, Precision: 0.1})
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Import(ctx, model.Observation{PeriodID: period.ID, FromPoint: "f", ToPoint: "p", Distance: 12.1, Precision: 0.1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Compute(ctx, period.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.WithdrawObservation(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.WithdrawObservation(ctx, second.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Publish(ctx, period.ID); err == nil {
		t.Fatal("expected publish to be refused after all observations were withdrawn, but it succeeded")
	}
}

// TestPublishStillAllowedWhenObservationsRemain ensures the guard does not
// over-block: a period whose latest result was computed from usable observations
// can still be published when those observations remain usable.
func TestPublishStillAllowedWhenObservationsRemain(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "publish-ok.db"))
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
	for _, point := range []model.Point{
		{ID: "f", NetworkID: network.ID, Role: model.PointFixed},
		{ID: "p", NetworkID: network.ID, Role: model.PointEstimated},
	} {
		if _, err := service.AddPoint(ctx, point); err != nil {
			t.Fatal(err)
		}
	}
	period, err := service.CreatePeriod(ctx, network.ID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Import(ctx, model.Observation{PeriodID: period.ID, FromPoint: "f", ToPoint: "p", Distance: 12, Precision: 0.1}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Compute(ctx, period.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Publish(ctx, period.ID); err != nil {
		t.Fatalf("expected publish to succeed with usable observations present: %v", err)
	}
}
