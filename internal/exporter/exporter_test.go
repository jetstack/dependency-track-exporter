package exporter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/go-kit/log"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

func TestFetchProjects_Pagination(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	var wantProjects []dtrack.Project
	for i := 0; i < 468; i++ {
		wantProjects = append(wantProjects, dtrack.Project{
			UUID: uuid.New(),
		})
	}

	mux.HandleFunc("/api/v1/project", func(w http.ResponseWriter, r *http.Request) {
		pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
		if err != nil {
			t.Fatalf("unexpected error converting pageSize to int: %s", err)
		}
		pageNumber, err := strconv.Atoi(r.URL.Query().Get("pageNumber"))
		if err != nil {
			t.Fatalf("unexpected error converting pageNumber to int: %s", err)
		}
		w.Header().Set("X-Total-Count", strconv.Itoa(len(wantProjects)))
		w.Header().Set("Content-type", "application/json")
		var projects []dtrack.Project
		for i := 0; i < pageSize; i++ {
			idx := (pageSize * (pageNumber - 1)) + i
			if idx >= len(wantProjects) {
				break
			}
			projects = append(projects, wantProjects[idx])
		}
		json.NewEncoder(w).Encode(projects)
	})

	client, err := dtrack.NewClient(server.URL)
	if err != nil {
		t.Fatalf("unexpected error setting up client: %s", err)
	}

	e := &Exporter{
		Client: client,
	}

	gotProjects, err := e.fetchProjects(context.Background())
	if err != nil {
		t.Fatalf("unexpected error fetching projects: %s", err)
	}

	if diff := cmp.Diff(wantProjects, gotProjects); diff != "" {
		t.Errorf("unexpected projects:\n%s", diff)
	}
}

func TestFetchProjectsByTag_Pagination(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	wantProjects := []dtrack.Project{
		{UUID: uuid.New(), Name: "prod-project"},
	}

	mux.HandleFunc("/api/v1/project/tag/prod", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Total-Count", "1")
		w.Header().Set("Content-type", "application/json")
		json.NewEncoder(w).Encode(wantProjects)
	})

	client, _ := dtrack.NewClient(server.URL)
	e := &Exporter{
		Client:      client,
		ProjectTags: []string{"prod"},
	}

	gotProjects, err := e.fetchProjects(context.Background())
	if err != nil {
		t.Fatalf("unexpected error fetching projects: %s", err)
	}

	if diff := cmp.Diff(wantProjects, gotProjects); diff != "" {
		t.Errorf("unexpected projects:\n%s", diff)
	}
}

func TestFetchPolicyViolations_Pagination(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	var wantPolicyViolations []dtrack.PolicyViolation
	for i := 0; i < 468; i++ {
		wantPolicyViolations = append(wantPolicyViolations, dtrack.PolicyViolation{
			UUID: uuid.New(),
		})
	}

	mux.HandleFunc("/api/v1/violation", func(w http.ResponseWriter, r *http.Request) {
		pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
		if err != nil {
			t.Fatalf("unexpected error converting pageSize to int: %s", err)
		}
		pageNumber, err := strconv.Atoi(r.URL.Query().Get("pageNumber"))
		if err != nil {
			t.Fatalf("unexpected error converting pageNumber to int: %s", err)
		}
		w.Header().Set("X-Total-Count", strconv.Itoa(len(wantPolicyViolations)))
		w.Header().Set("Content-type", "application/json")
		var policyViolations []dtrack.PolicyViolation
		for i := 0; i < pageSize; i++ {
			idx := (pageSize * (pageNumber - 1)) + i
			if idx >= len(wantPolicyViolations) {
				break
			}
			policyViolations = append(policyViolations, wantPolicyViolations[idx])
		}
		json.NewEncoder(w).Encode(policyViolations)
	})

	client, err := dtrack.NewClient(server.URL)
	if err != nil {
		t.Fatalf("unexpected error setting up client: %s", err)
	}

	e := &Exporter{
		Client: client,
	}

	gotPolicyViolations, err := e.fetchPolicyViolations(context.Background())
	if err != nil {
		t.Fatalf("unexpected error fetching projects: %s", err)
	}

	if diff := cmp.Diff(wantPolicyViolations, gotPolicyViolations); diff != "" {
		t.Errorf("unexpected policy violations:\n%s", diff)
	}
}

func TestExporter_Run(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	// Mock Portfolio metrics
	mux.HandleFunc("/api/v1/metrics/portfolio/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dtrack.PortfolioMetrics{})
	})

	// Mock Projects
	mux.HandleFunc("/api/v1/project", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Total-Count", "0")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]dtrack.Project{})
	})

	// Mock Violations
	mux.HandleFunc("/api/v1/violation", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Total-Count", "0")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]dtrack.PolicyViolation{})
	})

	client, _ := dtrack.NewClient(server.URL)
	e := &Exporter{
		Client: client,
		Logger: log.NewNopLogger(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start exporter in background with short interval
	go e.Run(ctx, 100*time.Millisecond)

	// Wait for at least one poll to complete
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		e.mutex.RLock()
		reg := e.registry
		e.mutex.RUnlock()
		if reg != nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatal("Exporter failed to populate registry in time")
}
