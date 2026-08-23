package service

import (
	"context"
	"deformation/internal/model"
	"deformation/internal/store"
	"path/filepath"
	"testing"
)

func TestBug04_InputDoesNotPersistOutOfRangePoint(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "coordinates.db"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	service := New(db)
	network, err := service.CreateNetwork(context.Background(), "site")
	if err != nil { t.Fatal(err) }
	if _, err = service.AddPointInput(context.Background(), network.ID, model.PointInput{ID: "bad", Role: model.PointEstimated, X: 100_000_001}); err == nil {
		t.Fatal("out-of-range coordinate was persisted")
	}
}
