package dependencytrack

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// TestGetProjects tests listing projects
func TestGetProjects(t *testing.T) {
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
			    "lastBomImport": %d,
			    "metrics": {
			      "critical": 0,
			      "high": 1,
			      "low": 2,
			      "medium": 3,
			      "unassigned": 4,
			      "inheritedRiskScore": 1240
			    },
			    "uuid": "fd1b10b9-678d-4af9-ad8e-877d1f357b03"
			  },
			  {
			    "name": "bar",
			    "version": "foo",
			    "active": false,
			    "metrics": {
			      "critical": 50,
			      "high": 25,
			      "low": 12,
			      "medium": 6,
			      "unassigned": 3,
			      "inheritedRiskScore": 2560.26
			    },
			    "uuid": "9b9a702a-a8b4-49fb-bb99-c05c1a8c8d49"
			  }
			]
			`,
			now.Unix(),
		)
	})

	got, err := client.GetProjects()
	if err != nil {
		t.Errorf("GetProjects returned error: %v", err)
	}

	want := []*Project{
		{
			Name:          "foo",
			Version:       "bar",
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
			UUID: "fd1b10b9-678d-4af9-ad8e-877d1f357b03",
		},
		{
			Name:          "bar",
			Version:       "foo",
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
			UUID: "9b9a702a-a8b4-49fb-bb99-c05c1a8c8d49",
		},
	}

	if !cmp.Equal(got, want) {
		t.Errorf("Got projects %v, want %v", got, want)
	}
}
