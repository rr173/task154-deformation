package model

import "fmt"

func (s PeriodStatus) CanAcceptObservation() bool {
	return s.IsOpen() && (s == PeriodEntering || s == PeriodPending || s == PeriodFailed || s == PeriodSucceeded)
}

func (s PeriodStatus) IsOpen() bool {
	return s != PeriodPublished
}

func (s PeriodStatus) CanCompute() bool {
	return s == PeriodEntering || s == PeriodPending || s == PeriodFailed || s == PeriodSucceeded
}

func (s PeriodStatus) CanPublish() bool {
	return s == PeriodSucceeded
}

func (s PeriodStatus) IsTerminal() bool {
	return s == PeriodPublished
}

func (s ObservationStatus) IsIncluded() bool {
	return s == ObservationValid || s == ObservationOutlier
}

// CanAddData reports whether the network still accepts new data ingestion
// (points, observation periods) and may be (re)marked ready. Archived networks
// are irrevocably closed to data ingestion, so they never satisfy this.
func (s NetworkStatus) CanAddData() bool {
	return s == NetworkDraft || s == NetworkReady
}

// IsArchived reports whether the network has been archived. Archiving is a
// terminal, irreversible closure of data ingestion: an archived network can no
// longer be marked ready, accept points, or open new observation periods.
func (s NetworkStatus) IsArchived() bool {
	return s == NetworkArchived
}

func (s NetworkStatus) CanArchive() bool {
	return s == NetworkDraft || s == NetworkReady || s == NetworkPublished
}

func ValidateTransition(before, after PeriodStatus) error {
	if before == after {
		return nil
	}
	allowed := map[PeriodStatus]map[PeriodStatus]bool{
		PeriodEntering:  {PeriodPending: true, PeriodRunning: true, PeriodFailed: true},
		PeriodPending:   {PeriodRunning: true, PeriodFailed: true},
		PeriodRunning:   {PeriodSucceeded: true, PeriodFailed: true},
		PeriodSucceeded: {PeriodPending: true, PeriodRunning: true, PeriodPublished: true},
		PeriodFailed:    {PeriodPending: true, PeriodRunning: true},
	}
	if allowed[before][after] {
		return nil
	}
	return fmt.Errorf("period transition %s -> %s is not allowed", before, after)
}
