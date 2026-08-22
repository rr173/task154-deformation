package service

import (
	"context"
	"github.com/rr173/task154-deformation/internal/model"
	"github.com/rr173/task154-deformation/internal/store"
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
