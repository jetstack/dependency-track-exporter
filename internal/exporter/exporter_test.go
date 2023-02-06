package exporter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	dtrack "github.com/DependencyTrack/client-go"
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
