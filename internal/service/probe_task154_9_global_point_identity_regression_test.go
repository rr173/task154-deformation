package service

import (
	"context"
	"deformation/internal/model"
	"deformation/internal/store"
	"path/filepath"
	"testing"
)

func TestBug09_PointIdentifierCannotMoveBetweenNetworks(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "identity.db"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	service := New(db)
	ctx := context.Background()
	first, err := service.CreateNetwork(ctx, "first")
	if err != nil { t.Fatal(err) }
	second, err := service.CreateNetwork(ctx, "second")
	if err != nil { t.Fatal(err) }
	if _, err = service.AddPointInput(ctx, first.ID, model.PointInput{ID: "station", Role: model.PointFixed}); err != nil { t.Fatal(err) }
	if _, err = service.AddPointInput(ctx, second.ID, model.PointInput{ID: "station", Role: model.PointEstimated}); err == nil {
		t.Fatal("point identifier was reassigned to another network")
	}
}
