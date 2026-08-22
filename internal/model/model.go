package model

import "time"

type NetworkStatus string

const (
	NetworkDraft     NetworkStatus = "draft"
	NetworkReady     NetworkStatus = "ready"
	NetworkPublished NetworkStatus = "published"
	NetworkArchived  NetworkStatus = "archived"
)

type PointRole string

const (
	PointFixed     PointRole = "fixed"
	PointEstimated PointRole = "estimated"
	PointDisabled  PointRole = "disabled"
)

type PeriodStatus string

const (
	PeriodEntering  PeriodStatus = "entering"
	PeriodPending   PeriodStatus = "pending"
	PeriodRunning   PeriodStatus = "running"
	PeriodSucceeded PeriodStatus = "succeeded"
	PeriodFailed    PeriodStatus = "failed"
	PeriodPublished PeriodStatus = "published"
)

type ObservationStatus string

const (
	ObservationValid     ObservationStatus = "valid"
	ObservationDuplicate ObservationStatus = "duplicate"
	ObservationOutlier   ObservationStatus = "outlier"
	ObservationWithdrawn ObservationStatus = "withdrawn"
)

type Network struct {
	ID, Name        string
	Status          NetworkStatus
	Description     string
	CoordinateDatum string
	CreatedAt       time.Time
}
type Point struct {
	ID, NetworkID string
	Role          PointRole
	X, Y, Z       float64
	ValidFrom     time.Time
	Label         string
}
type Period struct {
	ID, NetworkID string
	ObservedAt    time.Time
	Status        PeriodStatus
	Revision      int
	Note          string
}
type Observation struct {
	ID, PeriodID, FromPoint, ToPoint string
	Distance, Precision, Residual    float64
	Status                           ObservationStatus
	CreatedAt                        time.Time
	Source                           string
}
type Result struct {
	ID, PeriodID     string
	Version          int
	Published        bool
	Points           []PointResult
	CreatedAt        time.Time
	ObservationCount int
	MaxResidual      float64
}
type PointResult struct {
	PointID         string
	X, Y, Z, StdDev float64
	Deformed        bool
	Residual        float64
}
type Compare struct {
	PointID              string
	DX, DY, DZ, Distance float64
	Significant          bool
}

type NetworkSummary struct {
	Network          Network
	PointCount       int
	PeriodCount      int
	PublishedPeriods int
	LatestPeriod     *Period
}

type PeriodSummary struct {
	Period                Period
	ObservationCount      int
	ValidObservations     int
	OutlierObservations   int
	WithdrawnObservations int
	LatestResult          *Result
}

type RecoveryReport struct {
	RecoveredPeriodIDs []string
	FailedPeriodIDs    []string
	CheckedAt          time.Time
}

type ObservationIssue struct {
	ObservationID string
	Reason        string
	Magnitude     float64
}
