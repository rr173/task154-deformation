package main

import (
	"context"
	"flag"
	"fmt"
	"deformation/internal/httpapi"
	"deformation/internal/model"
	"deformation/internal/service"
	"deformation/internal/store"
	"deformation/internal/webui"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	db := flag.String("db", "deformation.db", "sqlite database")
	addr := flag.String("addr", ":8080", "listen address")
	smoke := flag.Bool("smoke-test", false, "run self test")
	flag.Parse()
	if *smoke {
		runSmoke()
		return
	}
	s, e := store.Open(*db)
	if e != nil {
		log.Fatal(e)
	}
	defer s.Close()
	core := service.New(s)
	mux := http.NewServeMux()
	api := httpapi.New(core)
	mux.Handle("/v1/", api.Handler())
	mux.Handle("/", webui.Handler())
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"ok":true}`) })
	mux.HandleFunc("POST /v1/demo", func(w http.ResponseWriter, r *http.Request) {
		if e := seed(r.Context(), core); e != nil {
			http.Error(w, e.Error(), 400)
			return
		}
		fmt.Fprint(w, `{"ok":true}`)
	})
	log.Fatal(http.ListenAndServe(*addr, mux))
}
func seed(ctx context.Context, s *service.Service) error {
	n, e := s.CreateNetwork(ctx, "demo")
	if e != nil {
		return e
	}
	for _, p := range []model.Point{{ID: "fixed", NetworkID: n.ID, Role: model.PointFixed, X: 0}, {ID: "target", NetworkID: n.ID, Role: model.PointEstimated}} {
		if _, e = s.AddPoint(ctx, p); e != nil {
			return e
		}
	}
	period, e := s.CreatePeriod(ctx, n.ID, time.Now())
	if e != nil {
		return e
	}
	_, e = s.Import(ctx, model.Observation{PeriodID: period.ID, FromPoint: "fixed", ToPoint: "target", Distance: 10, Precision: 0.01})
	if e != nil {
		return e
	}
	_, e = s.Compute(ctx, period.ID)
	return e
}
func runSmoke() {
	dir, e := os.MkdirTemp("", "deformation-smoke")
	if e != nil {
		log.Fatal(e)
	}
	defer os.RemoveAll(dir)
	db, e := store.Open(filepath.Join(dir, "s.db"))
	if e != nil {
		log.Fatal(e)
	}
	defer db.Close()
	if e = seed(context.Background(), service.New(db)); e != nil {
		log.Fatal(e)
	}
	fmt.Println("smoke-test passed")
}
