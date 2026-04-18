package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/MMahesa/incident-api/internal/incidents"
)

func TestIncidentLifecycle(t *testing.T) {
	store, err := incidents.NewFileStore(filepath.Join(t.TempDir(), "incidents.json"))
	if err != nil {
		t.Fatal(err)
	}
	handler := NewServer(store)

	createBody := map[string]any{
		"title":       "Database saturation",
		"service":     "postgres-primary",
		"severity":    "critical",
		"status":      "open",
		"description": "Connections exceeded threshold",
		"owner":       "backend-team",
	}
	payload, _ := json.Marshal(createBody)

	request := httptest.NewRequest(http.MethodPost, "/v1/incidents", bytes.NewReader(payload))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}

	secondBody := map[string]any{
		"title":       "Queue delay",
		"service":     "job-runner",
		"severity":    "medium",
		"status":      "open",
		"description": "Queue processing delay increased",
		"owner":       "platform-team",
	}
	payload, _ = json.Marshal(secondBody)
	request = httptest.NewRequest(http.MethodPost, "/v1/incidents", bytes.NewReader(payload))
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201 for second create, got %d", recorder.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/v1/incidents?status=open&limit=999&offset=0&sort=oldest", nil)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"sort":"oldest"`)) {
		t.Fatal("expected sort metadata in response")
	}

	request = httptest.NewRequest(http.MethodGet, "/v1/incidents/1", nil)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 on get by id, got %d", recorder.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/v1/incidents/stats", nil)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 on stats, got %d", recorder.Code)
	}

	request = httptest.NewRequest(http.MethodDelete, "/v1/incidents/1", nil)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete, got %d", recorder.Code)
	}
}
