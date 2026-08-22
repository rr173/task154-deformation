package model

import (
	"testing"
	"time"
)

func TestInputsAndTransitions(t *testing.T) {
	if err := (NetworkInput{Name: "A"}).Validate(); err == nil {
		t.Fatal("short network name was accepted")
	}
	if err := (PointInput{ID: "p", Role: PointEstimated}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (ObservationInput{FromPoint: "a", ToPoint: "b", Distance: 1, Precision: 0.1}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := ValidateTransition(PeriodPending, PeriodRunning); err != nil {
		t.Fatal(err)
	}
	if err := ValidateTransition(PeriodPublished, PeriodPending); err == nil {
		t.Fatal("terminal period became mutable")
	}
	when := (PeriodInput{}).Normalize(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	if when.IsZero() {
		t.Fatal("period time was not normalized")
	}
}
