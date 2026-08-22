package service

import (
	"context"
	"deformation/internal/model"
	"deformation/internal/store"
	"math"
	"path/filepath"
	"testing"
)

func TestBug08_FieldImportRejectsNonFiniteMeasurements(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "nonfinite.db"))
	if err != nil { t.Fatal(err) }
	defer db.Close()
	service := New(db)
	ctx := context.Background()
	network, err := service.CreateNetwork(ctx, "site")
	if err != nil { t.Fatal(err) }
	for _, point := range []model.Point{{ID: "fixed", NetworkID: network.ID, Role: model.PointFixed}, {ID: "target", NetworkID: network.ID, Role: model.PointEstimated}} {
		if _, err = service.AddPoint(ctx, point); err != nil { t.Fatal(err) }
	}
	period, err := service.CreatePeriod(ctx, network.ID, service.now())
	if err != nil { t.Fatal(err) }
	if _, err = service.ImportInput(ctx, period.ID, model.ObservationInput{FromPoint: "fixed", ToPoint: "target", Distance: math.NaN(), Precision: 0.1, Source: "field-import"}); err == nil {
		t.Fatal("field import accepted a non-finite distance")
	}
}
