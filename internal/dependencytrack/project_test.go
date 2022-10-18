package dependencytrack

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// TestGetAllProjects tests listing all projects
func TestGetAllProjects(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	now := time.Now().Truncate(time.Second)

	mux.HandleFunc("/api/v1/project", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodGet {
			t.Errorf("Got request method %v, want %v", got, http.MethodGet)
		}
		fmt.Fprintf(w,
			`
			[
			  {
			    "name": "foo",
			    "version": "bar",
			    "active": true,
			    "classifier": "CONTAINER",
			    "lastBomImport": %d,
			    "metrics": {
			      "critical": 0,
			      "high": 1,
			      "low": 2,
			      "medium": 3,
			      "unassigned": 4,
			      "inheritedRiskScore": 1240
			    },
			    "tags": [
			     {
			       "name": "foo"
			     },
			     {
			       "name": "bar"
			     }
			    ],
			    "uuid": "fd1b10b9-678d-4af9-ad8e-877d1f357b03"
			  },
			  {
			    "name": "bar",
			    "version": "foo",
			    "active": false,
			    "classifier": "APPLICATION",
			    "metrics": {
			      "critical": 50,
			      "high": 25,
			      "low": 12,
			      "medium": 6,
			      "unassigned": 3,
			      "inheritedRiskScore": 2560.26
			    },
			    "tags": [
			     {
			       "name": "foobar"
			     }
			    ],
			    "uuid": "9b9a702a-a8b4-49fb-bb99-c05c1a8c8d49"
			  }
			]
			`,
			now.Unix(),
		)
	})

	got, err := client.GetAllProjects()
	if err != nil {
		t.Errorf("GetProjects returned error: %v", err)
	}

	want := []*Project{
		{
			Name:          "foo",
			Version:       "bar",
			Classifier:    "CONTAINER",
			Active:        true,
			LastBomImport: Time{now},
			Metrics: ProjectMetrics{
				Critical:           0,
				High:               1,
				Low:                2,
				Medium:             3,
				Unassigned:         4,
				InheritedRiskScore: 1240,
			},
			Tags: []ProjectTag{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
			},
			UUID: "fd1b10b9-678d-4af9-ad8e-877d1f357b03",
		},
		{
			Name:          "bar",
			Version:       "foo",
			Classifier:    "APPLICATION",
			Active:        false,
			LastBomImport: Time{},
			Metrics: ProjectMetrics{
				Critical:           50,
				High:               25,
				Low:                12,
				Medium:             6,
				Unassigned:         3,
				InheritedRiskScore: 2560.26,
			},
			Tags: []ProjectTag{
				{
					Name: "foobar",
				},
			},
			UUID: "9b9a702a-a8b4-49fb-bb99-c05c1a8c8d49",
		},
	}

	if !cmp.Equal(got, want) {
		t.Errorf("Got projects %v, want %v", got, want)
	}
}

// TestGetProjectsByPage tests listing projects by page
func TestGetProjectsByPage(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	now := time.Now().Truncate(time.Second)

	mux.HandleFunc("/api/v1/project", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodGet {
			t.Errorf("Got request method %v, want %v", got, http.MethodGet)
		}
		if r.Header.Get("pageNumber") == "1" && r.Header.Get("pageSize") == "1" {
			fmt.Fprintf(w,
				`
			[
			  {
			    "name": "foo",
			    "version": "bar",
			    "active": true,
			    "classifier": "CONTAINER",
			    "lastBomImport": %d,
			    "metrics": {
			      "critical": 0,
			      "high": 1,
			      "low": 2,
			      "medium": 3,
			      "unassigned": 4,
			      "inheritedRiskScore": 1240
			    },
			    "tags": [
			     {
			       "name": "foo"
			     },
			     {
			       "name": "bar"
			     }
			    ],
			    "uuid": "fd1b10b9-678d-4af9-ad8e-877d1f357b03"
			  }
			]
			`,
				now.Unix(),
			)
		}
		if r.Header.Get("pageNumber") == "2" && r.Header.Get("pageSize") == "1" {
			fmt.Fprintf(w,
				`
			[
			  {
			    "name": "bar",
			    "version": "foo",
			    "active": false,
			    "classifier": "APPLICATION",
				"lastBomImport": %d,
			    "metrics": {
			      "critical": 50,
			      "high": 25,
			      "low": 12,
			      "medium": 6,
			      "unassigned": 3,
			      "inheritedRiskScore": 2560.26
			    },
			    "tags": [
			     {
			       "name": "foobar"
			     }
			    ],
			    "uuid": "9b9a702a-a8b4-49fb-bb99-c05c1a8c8d49"
			  }
			]
			`,
				now.Unix(),
			)
		}
	})

	gotPageOne, err := client.GetProjects(1, 1)
	if err != nil {
		t.Errorf("GetProjects returned error: %v", err)
	}

	wantPageOne := []*Project{
		{
			Name:          "foo",
			Version:       "bar",
			Classifier:    "CONTAINER",
			Active:        true,
			LastBomImport: Time{now},
			Metrics: ProjectMetrics{
				Critical:           0,
				High:               1,
				Low:                2,
				Medium:             3,
				Unassigned:         4,
				InheritedRiskScore: 1240,
			},
			Tags: []ProjectTag{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
			},
			UUID: "fd1b10b9-678d-4af9-ad8e-877d1f357b03",
		},
	}

	if !cmp.Equal(gotPageOne, wantPageOne) {
		t.Errorf("Got projects %v, wantPageOne %v", gotPageOne, wantPageOne)
	}

	gotPageTwo, err := client.GetProjects(2, 1)
	if err != nil {
		t.Errorf("GetProjects returned error: %v", err)
	}

	wantPageTwo := []*Project{
		{
			Name:          "bar",
			Version:       "foo",
			Classifier:    "APPLICATION",
			Active:        false,
			LastBomImport: Time{now},
			Metrics: ProjectMetrics{
				Critical:           50,
				High:               25,
				Low:                12,
				Medium:             6,
				Unassigned:         3,
				InheritedRiskScore: 2560.26,
			},
			Tags: []ProjectTag{
				{
					Name: "foobar",
				},
			},
			UUID: "9b9a702a-a8b4-49fb-bb99-c05c1a8c8d49",
		},
	}

	if !cmp.Equal(gotPageTwo, wantPageTwo) {
		t.Errorf("Got projects %v, wantPageOne %v", gotPageOne, wantPageOne)
	}
}
