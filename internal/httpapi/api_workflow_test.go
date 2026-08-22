package httpapi

import (
	"bytes"
	"encoding/json"
	"github.com/rr173/task154-deformation/internal/service"
	"github.com/rr173/task154-deformation/internal/store"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestWorkflowAPI(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	handler := New(service.New(db)).Handler()
	request := func(method, path, body string) map[string]any {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(method, path, bytes.NewBufferString(body)))
		if recorder.Code < 200 || recorder.Code > 299 {
			t.Fatalf("%s %s: %d %s", method, path, recorder.Code, recorder.Body.String())
		}
		var value map[string]any
		if err := json.Unmarshal(recorder.Body.Bytes(), &value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	network := request(http.MethodPost, "/v1/networks", `{"name":"site","datum":"local"}`)
	networkID := network["ID"].(string)
	request(http.MethodPost, "/v1/networks/"+networkID+"/points", `{"id":"fixed","role":"fixed","x":0}`)
	request(http.MethodPost, "/v1/networks/"+networkID+"/points", `{"id":"target","role":"estimated","x":0}`)
	period := request(http.MethodPost, "/v1/networks/"+networkID+"/periods", `{}`)
	periodID := period["ID"].(string)
	request(http.MethodPost, "/v1/periods/"+periodID+"/observations", `{"from_point":"fixed","to_point":"target","distance":10,"precision":0.1}`)
	result := request(http.MethodPost, "/v1/periods/"+periodID+"/compute", "")
	if result["ObservationCount"].(float64) != 1 {
		t.Fatal(result)
	}
}
