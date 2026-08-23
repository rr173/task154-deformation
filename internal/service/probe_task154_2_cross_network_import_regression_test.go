package service

import (
	"context"
	"deformation/internal/model"
	"deformation/internal/store"
	"path/filepath"
	"testing"
)

func TestBug02_ImportRejectsOneForeignEndpoint(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "cross-network.db"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	service := New(db)
	ctx := context.Background()
	first, err := service.CreateNetwork(ctx, "first")
	if err != nil { t.Fatal(err) }
	second, err := service.CreateNetwork(ctx, "second")
	if err != nil { t.Fatal(err) }
	if _, err = service.AddPoint(ctx, model.Point{ID: "fixed", NetworkID: first.ID, Role: model.PointFixed}); err != nil { t.Fatal(err) }
	if _, err = service.AddPoint(ctx, model.Point{ID: "foreign", NetworkID: second.ID, Role: model.PointEstimated}); err != nil { t.Fatal(err) }
	period, err := service.CreatePeriod(ctx, first.ID, service.now())
	if err != nil { t.Fatal(err) }
	if _, err = service.ImportInput(ctx, period.ID, model.ObservationInput{FromPoint: "fixed", ToPoint: "foreign", Distance: 10, Precision: 0.1}); err == nil {
		t.Fatal("observation with one foreign endpoint was accepted")
	}
}
