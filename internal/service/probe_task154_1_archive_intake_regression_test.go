package service

import (
	"context"
	"deformation/internal/model"
	"deformation/internal/store"
	"path/filepath"
	"testing"
	"time"
)

func TestBug01_ArchivedNetworkRejectsEveryIntakePath(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "archive.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	service := New(db)
	ctx := context.Background()
	network, err := service.CreateNetwork(ctx, "survey")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.AddPoint(ctx, model.Point{ID: "fixed", NetworkID: network.ID, Role: model.PointFixed}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.ArchiveNetwork(ctx, network.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = service.SetNetworkReady(ctx, network.ID); err == nil {
		t.Error("archived network became ready again")
	}
	if _, err = service.AddPoint(ctx, model.Point{ID: "new", NetworkID: network.ID, Role: model.PointEstimated}); err == nil {
		t.Error("archived network accepted a point")
	}
	if _, err = service.CreatePeriod(ctx, network.ID, time.Now()); err == nil {
		t.Error("archived network accepted an observation period")
	}
}
