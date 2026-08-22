package httpapi

import (
	"bytes"
	"github.com/rr173/task154-deformation/internal/service"
	"github.com/rr173/task154-deformation/internal/store"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestNetworkAPI(t *testing.T) {
	db, e := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	h := New(service.New(db)).Handler()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/v1/networks", bytes.NewBufferString(`{"name":"site"}`)))
	if w.Code != http.StatusCreated {
		t.Fatal(w.Code, w.Body.String())
	}
}
