package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/MMahesa/incident-api/internal/incidents"
)

type Server struct {
	store incidents.Store
}

type contextKey string

const requestIDKey contextKey = "request_id"

var requestSeq atomic.Int64

func NewServer(store incidents.Store) http.Handler {
	server := &Server{store: store}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", server.handleHealth)
	mux.HandleFunc("GET /v1/incidents", server.handleListIncidents)
	mux.HandleFunc("POST /v1/incidents", server.handleCreateIncident)
	mux.HandleFunc("PUT /v1/incidents/{id}", server.handleUpdateIncident)
	mux.HandleFunc("DELETE /v1/incidents/{id}", server.handleDeleteIncident)

	return withMiddleware(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	})
}

func (s *Server) handleListIncidents(w http.ResponseWriter, r *http.Request) {
	options := incidents.ParseListOptions(map[string]string{
		"service":  r.URL.Query().Get("service"),
		"status":   r.URL.Query().Get("status"),
		"severity": r.URL.Query().Get("severity"),
		"owner":    r.URL.Query().Get("owner"),
		"search":   r.URL.Query().Get("search"),
		"limit":    r.URL.Query().Get("limit"),
		"offset":   r.URL.Query().Get("offset"),
	})
	result, err := s.store.List(r.Context(), options)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": result.Items,
		"meta": map[string]int{
			"count":  len(result.Items),
			"total":  result.Total,
			"limit":  options.Limit,
			"offset": options.Offset,
		},
	})
}

func (s *Server) handleCreateIncident(w http.ResponseWriter, r *http.Request) {
	var input incidents.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}

	item, err := s.store.Create(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (s *Server) handleUpdateIncident(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid incident id"))
		return
	}

	var input incidents.UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}

	item, err := s.store.Update(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, incidents.ErrNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (s *Server) handleDeleteIncident(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid incident id"))
		return
	}

	if err := s.store.Delete(r.Context(), id); err != nil {
		if errors.Is(err, incidents.ErrNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "deleted"})
}

func withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := nextRequestID()
		r = r.WithContext(context.WithValue(r.Context(), requestIDKey, requestID))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
		log.Printf("request_id=%s method=%s path=%s duration=%s", requestID, r.Method, strings.TrimSpace(r.URL.Path), time.Since(start).Round(time.Millisecond))
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{
		"error": err.Error(),
	})
}

func nextRequestID() string {
	return "req-" + strconv.FormatInt(requestSeq.Add(1), 10)
}
