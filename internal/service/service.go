package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/rr173/task154-deformation/internal/adjustment"
	"github.com/rr173/task154-deformation/internal/model"
	"github.com/rr173/task154-deformation/internal/store"
	"sync"
	"time"
)

type Service struct {
	s   *store.Store
	mu  sync.Mutex
	now func() time.Time
}

func New(s *store.Store) *Service { return &Service{s: s, now: time.Now} }
func (v *Service) CreateNetwork(ctx context.Context, name string) (model.Network, error) {
	if name == "" {
		return model.Network{}, fmt.Errorf("name required")
	}
	n := model.Network{ID: uuid.NewString(), Name: name, Status: model.NetworkDraft, CoordinateDatum: "local-engineering", CreatedAt: v.now().UTC()}
	return n, v.s.SaveNetwork(ctx, n)
}
func (v *Service) AddPoint(ctx context.Context, p model.Point) (model.Point, error) {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	if p.Role == "" {
		p.Role = model.PointEstimated
	}
	network, e := v.s.Network(ctx, p.NetworkID)
	if e != nil {
		return p, e
	}
	if !network.Status.CanAddData() {
		return p, fmt.Errorf("network does not accept new points")
	}
	if p.Role != model.PointFixed && p.Role != model.PointEstimated && p.Role != model.PointDisabled {
		return p, fmt.Errorf("invalid point role")
	}
	if p.ValidFrom.IsZero() {
		p.ValidFrom = v.now().UTC()
	}
	return p, v.s.SavePoint(ctx, p)
}
func (v *Service) CreatePeriod(ctx context.Context, network string, at time.Time) (model.Period, error) {
	if _, e := v.s.Network(ctx, network); e != nil {
		return model.Period{}, e
	}
	if at.IsZero() {
		at = v.now()
	}
	p := model.Period{ID: uuid.NewString(), NetworkID: network, ObservedAt: at.UTC(), Status: model.PeriodEntering, Revision: 1}
	return p, v.s.SavePeriod(ctx, p)
}
func (v *Service) Import(ctx context.Context, o model.Observation) (model.Observation, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if o.Distance <= 0 || o.Precision <= 0 {
		return o, fmt.Errorf("distance and precision must be positive")
	}
	p, e := v.s.Period(ctx, o.PeriodID)
	if e != nil {
		return o, e
	}
	if !p.Status.CanAcceptObservation() {
		return o, fmt.Errorf("published period is immutable")
	}
	if _, e := v.s.Point(ctx, o.FromPoint); e != nil {
		return o, fmt.Errorf("source point: %w", e)
	}
	if _, e := v.s.Point(ctx, o.ToPoint); e != nil {
		return o, fmt.Errorf("target point: %w", e)
	}
	if o.ID == "" {
		o.ID = uuid.NewString()
	}
	o.Status = model.ObservationValid
	o.CreatedAt = v.now().UTC()
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s/%s/%.6f/%.6f", o.FromPoint, o.ToPoint, o.Distance, o.Precision)))
	e = v.s.SaveObservation(ctx, o, fmt.Sprintf("%x", sum))
	if e != nil {
		if e == sql.ErrNoRows {
			return o, e
		}
		o.Status = model.ObservationDuplicate
		return o, nil
	}
	if err := model.ValidateTransition(p.Status, model.PeriodPending); err != nil {
		return o, err
	}
	p.Status = model.PeriodPending
	p.Revision++
	_ = v.s.UpdatePeriod(ctx, p)
	return o, nil
}
func (v *Service) Compute(ctx context.Context, id string) (model.Result, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	p, e := v.s.Period(ctx, id)
	if e != nil {
		return model.Result{}, e
	}
	if !p.Status.CanCompute() {
		return model.Result{}, fmt.Errorf("published period cannot be recalculated")
	}
	if err := model.ValidateTransition(p.Status, model.PeriodRunning); err != nil {
		return model.Result{}, err
	}
	p.Status = model.PeriodRunning
	_ = v.s.UpdatePeriod(ctx, p)
	points, e := v.s.Points(ctx, p.NetworkID)
	if e != nil {
		return model.Result{}, e
	}
	obs, e := v.s.Observations(ctx, id)
	if e != nil {
		return model.Result{}, e
	}
	out, e := adjustment.Solve(points, obs)
	if e != nil {
		p.Status = model.PeriodFailed
		_ = v.s.UpdatePeriod(ctx, p)
		return model.Result{}, e
	}
	prior, _ := v.s.Results(ctx, id)
	baseline, _ := v.previousPublished(ctx, p.NetworkID, p.ID)
	out = adjustment.MarkDeformation(out, baseline, 0.01)
	issues, maxResidual := adjustment.ClassifyResiduals(out, obs, 3*meanPrecision(obs))
	if len(issues) > 0 {
		_ = v.markOutliers(ctx, obs, issues)
	}
	r := model.Result{ID: uuid.NewString(), PeriodID: id, Version: len(prior) + 1, Points: adjustment.SortResults(out), CreatedAt: v.now().UTC(), ObservationCount: adjustment.ObservationCount(obs), MaxResidual: maxResidual}
	if e = v.s.SaveResult(ctx, r); e != nil {
		return r, e
	}
	if err := model.ValidateTransition(p.Status, model.PeriodSucceeded); err != nil {
		return r, err
	}
	p.Status = model.PeriodSucceeded
	p.Revision++
	_ = v.s.UpdatePeriod(ctx, p)
	return r, nil
}
func (v *Service) Publish(ctx context.Context, id string) (model.Result, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	p, e := v.s.Period(ctx, id)
	if e != nil {
		return model.Result{}, e
	}
	if e := v.ValidatePublication(ctx, id); e != nil {
		return model.Result{}, e
	}
	all, e := v.s.Results(ctx, id)
	if e != nil || len(all) == 0 {
		return model.Result{}, fmt.Errorf("no computed result")
	}
	r := all[len(all)-1]
	r.Published = true
	if e = v.s.SaveResult(ctx, r); e != nil {
		return r, e
	}
	if err := model.ValidateTransition(p.Status, model.PeriodPublished); err != nil {
		return r, err
	}
	p.Status = model.PeriodPublished
	p.Revision++
	_ = v.s.UpdatePeriod(ctx, p)
	return r, nil
}
func (v *Service) Compare(ctx context.Context, first, second string) ([]model.Compare, error) {
	a, e := v.s.Results(ctx, first)
	if e != nil || len(a) == 0 {
		return nil, fmt.Errorf("first result unavailable")
	}
	b, e := v.s.Results(ctx, second)
	if e != nil || len(b) == 0 {
		return nil, fmt.Errorf("second result unavailable")
	}
	return adjustment.Compare(a[len(a)-1].Points, b[len(b)-1].Points, 0.01), nil
}
