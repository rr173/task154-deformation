package httpapi

import (
	"encoding/json"
	"errors"
	"deformation/internal/model"
	"deformation/internal/service"
	"io"
	"net/http"
)

type API struct{ S *service.Service }

func New(s *service.Service) *API { return &API{S: s} }

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/networks", a.networks)
	mux.HandleFunc("POST /v1/networks", a.network)
	mux.HandleFunc("GET /v1/networks/{id}", a.networkDetail)
	mux.HandleFunc("GET /v1/networks/{id}/inspection", a.inspection)
	mux.HandleFunc("GET /v1/networks/{id}/coverage", a.coverage)
	mux.HandleFunc("GET /v1/networks/{id}/points", a.points)
	mux.HandleFunc("POST /v1/networks/{id}/points", a.point)
	mux.HandleFunc("POST /v1/networks/{id}/ready", a.ready)
	mux.HandleFunc("POST /v1/networks/{id}/archive", a.archive)
	mux.HandleFunc("GET /v1/networks/{id}/periods", a.periods)
	mux.HandleFunc("POST /v1/networks/{id}/periods", a.period)
	mux.HandleFunc("GET /v1/periods/{id}", a.periodDetail)
	mux.HandleFunc("GET /v1/periods/{id}/observations", a.observations)
	mux.HandleFunc("POST /v1/periods/{id}/observations", a.observation)
	mux.HandleFunc("POST /v1/periods/{id}/compute", a.compute)
	mux.HandleFunc("POST /v1/periods/{id}/publish", a.publish)
	mux.HandleFunc("GET /v1/periods/{id}/result", a.result)
	mux.HandleFunc("GET /v1/periods/{id}/quality", a.quality)
	mux.HandleFunc("POST /v1/observations/{id}/withdraw", a.withdraw)
	mux.HandleFunc("POST /v1/observations/{id}/restore", a.restore)
	mux.HandleFunc("GET /v1/compare", a.compare)
	mux.HandleFunc("POST /v1/recovery", a.recovery)
	return mux
}

func decode(w http.ResponseWriter, r *http.Request, value any) bool {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, errors.New("request must contain one JSON value"))
		return false
	}
	return true
}

func write(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeStatus(w, status, map[string]string{"error": err.Error()})
}

func writeStatus(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (a *API) networks(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.Networks(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	write(w, value)
}

func (a *API) network(w http.ResponseWriter, r *http.Request) {
	var input model.NetworkInput
	if !decode(w, r, &input) {
		return
	}
	value, err := a.S.CreateNetworkInput(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeStatus(w, http.StatusCreated, value)
}

func (a *API) networkDetail(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.Network(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	write(w, value)
}

func (a *API) inspection(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.InspectNetwork(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	write(w, value)
}

func (a *API) coverage(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.InspectPointCoverage(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	write(w, value)
}

func (a *API) points(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.Points(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	write(w, value)
}

func (a *API) point(w http.ResponseWriter, r *http.Request) {
	var input model.PointInput
	if !decode(w, r, &input) {
		return
	}
	value, err := a.S.AddPointInput(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeStatus(w, http.StatusCreated, value)
}

func (a *API) ready(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.SetNetworkReady(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	write(w, value)
}

func (a *API) archive(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.ArchiveNetwork(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	write(w, value)
}

func (a *API) periods(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.Periods(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	write(w, value)
}

func (a *API) period(w http.ResponseWriter, r *http.Request) {
	var input model.PeriodInput
	if !decode(w, r, &input) {
		return
	}
	value, err := a.S.CreatePeriodInput(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeStatus(w, http.StatusCreated, value)
}

func (a *API) periodDetail(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.PeriodSummary(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	write(w, value)
}

func (a *API) observations(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.Observations(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	write(w, value)
}

func (a *API) observation(w http.ResponseWriter, r *http.Request) {
	var input model.ObservationInput
	if !decode(w, r, &input) {
		return
	}
	value, err := a.S.ImportInput(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeStatus(w, http.StatusCreated, value)
}

func (a *API) compute(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.Compute(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	write(w, value)
}

func (a *API) publish(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.Publish(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	write(w, value)
}

func (a *API) result(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.Result(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	write(w, value)
}

func (a *API) quality(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.QualityIssues(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	write(w, value)
}

func (a *API) withdraw(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.WithdrawObservation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	write(w, value)
}

func (a *API) restore(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.RestoreObservation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	write(w, value)
}

func (a *API) compare(w http.ResponseWriter, r *http.Request) {
	first, second := r.URL.Query().Get("first"), r.URL.Query().Get("second")
	if first == "" || second == "" {
		writeError(w, http.StatusBadRequest, errors.New("first and second periods are required"))
		return
	}
	value, err := a.S.Compare(r.Context(), first, second)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	write(w, value)
}

func (a *API) recovery(w http.ResponseWriter, r *http.Request) {
	value, err := a.S.Recover(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	write(w, value)
}
