package model

import (
	"fmt"
	"strings"
	"time"
)

type NetworkInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Datum       string `json:"datum"`
}

type PointInput struct {
	ID    string    `json:"id"`
	Role  PointRole `json:"role"`
	X     float64   `json:"x"`
	Y     float64   `json:"y"`
	Z     float64   `json:"z"`
	Label string    `json:"label"`
}

type PeriodInput struct {
	ObservedAt time.Time `json:"observed_at"`
	Note       string    `json:"note"`
}

type ObservationInput struct {
	ID        string  `json:"id"`
	FromPoint string  `json:"from_point"`
	ToPoint   string  `json:"to_point"`
	Distance  float64 `json:"distance"`
	Precision float64 `json:"precision"`
	Source    string  `json:"source"`
}

func (v NetworkInput) Validate() error {
	if len(strings.TrimSpace(v.Name)) < 2 {
		return fmt.Errorf("network name must contain at least two characters")
	}
	if len(v.Name) > 120 || len(v.Description) > 1000 || len(v.Datum) > 120 {
		return fmt.Errorf("network text fields are too long")
	}
	return nil
}

func (v PointInput) Validate() error {
	if len(strings.TrimSpace(v.ID)) == 0 {
		return fmt.Errorf("point id is required")
	}
	if len(v.ID) > 80 || len(v.Label) > 160 {
		return fmt.Errorf("point id or label is too long")
	}
	switch v.Role {
	case PointFixed, PointEstimated, PointDisabled:
		return nil
	default:
		return fmt.Errorf("unsupported point role %q", v.Role)
	}
}

func (v PeriodInput) Normalize(now time.Time) time.Time {
	if v.ObservedAt.IsZero() {
		return now.UTC()
	}
	return v.ObservedAt.UTC()
}

func (v ObservationInput) Validate() error {
	if v.FromPoint == "" || v.ToPoint == "" || v.FromPoint == v.ToPoint {
		return fmt.Errorf("observation must connect two distinct points")
	}
	if v.Distance <= 0 || v.Precision <= 0 {
		return fmt.Errorf("distance and precision must be positive")
	}
	if v.Distance > 1_000_000 || v.Precision > 1000 {
		return fmt.Errorf("observation values exceed supported range")
	}
	return nil
}
