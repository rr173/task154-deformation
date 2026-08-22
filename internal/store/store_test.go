package store

import (
	"context"
	"deformation/internal/model"
	"path/filepath"
	"testing"
)

func TestCatalogRoundTrip(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "store.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	network := model.Network{ID: "n1", Name: "network", Status: model.NetworkDraft}
	if err := s.SaveNetwork(ctx, network); err != nil {
		t.Fatal(err)
	}
	if err := s.SavePoint(ctx, model.Point{ID: "p1", NetworkID: "n1", Role: model.PointFixed}); err != nil {
		t.Fatal(err)
	}
	if err := s.SavePeriod(ctx, model.Period{ID: "r1", NetworkID: "n1", Status: model.PeriodEntering}); err != nil {
		t.Fatal(err)
	}
	networks, err := s.Networks(ctx)
	if err != nil || len(networks) != 1 || networks[0].Name != "network" {
		t.Fatal(networks, err)
	}
	points, err := s.Points(ctx, "n1")
	if err != nil || len(points) != 1 {
		t.Fatal(points, err)
	}
}
