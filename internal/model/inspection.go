package model

import "time"

type NetworkInspection struct {
	NetworkID           string
	InspectedAt         time.Time
	Ready               bool
	FixedPointCount     int
	EstimatedPointCount int
	DisabledPointCount  int
	ConnectedPointCount int
	ObservationCount    int
	MissingPointIDs     []string
	ReadinessWarnings   []string
	Periods             []PeriodInspection
}

type PeriodInspection struct {
	PeriodID         string
	Status           PeriodStatus
	ObservedAt       time.Time
	ObservationCount int
	IncludedCount    int
	WithdrawnCount   int
	HasResult        bool
	Published        bool
	Warnings         []string
}

type PointCoverage struct {
	PointID          string
	Role             PointRole
	ObservationCount int
	Connected        bool
	LatestPeriodID   string
}

func (v NetworkInspection) IsHealthy() bool {
	return v.FixedPointCount > 0 &&
		v.EstimatedPointCount > 0 &&
		len(v.MissingPointIDs) == 0 &&
		len(v.ReadinessWarnings) == 0
}
