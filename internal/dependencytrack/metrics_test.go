package dependencytrack

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestGetCurrentPortfolioMetrics tests getting current portfolio metrics
func TestGetCurrentPortfolioMetrics(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/v1/metrics/portfolio/current", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodGet {
			t.Errorf("Got request method %v, want %v", got, http.MethodGet)
		}
		fmt.Fprintf(w,
			`
			{
			  "critical": 0,
			  "high": 1,
			  "low": 2,
			  "medium": 3,
			  "unassigned": 4,
			  "inheritedRiskScore": 2500.42,
			  "findingsAudited": 15,
			  "findingsUnaudited": 134
			}
			`,
		)
	})

	got, err := client.GetCurrentPortfolioMetrics()
	if err != nil {
		t.Errorf("GetCurrentPortfolioMetrics returned error: %v", err)
	}

	want := &PortfolioMetrics{
		Critical:           0,
		High:               1,
		Low:                2,
		Medium:             3,
		Unassigned:         4,
		InheritedRiskScore: 2500.42,
		FindingsAudited:    15,
		FindingsUnaudited:  134,
	}

	if !cmp.Equal(got, want) {
		t.Errorf("Got portfolio metrics %v, want %v", got, want)
	}
}
