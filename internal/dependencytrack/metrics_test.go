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
			  "inheritedRiskScore": 2500.42
			}
			`,
		)
	})

	got, err := client.GetCurrentPortfolioMetrics()
	if err != nil {
		t.Errorf("GetCurrentPortfolioMetrics returned error: %v", err)
	}

	want := &PortfolioMetrics{
		InheritedRiskScore: 2500.42,
	}

	if !cmp.Equal(got, want) {
		t.Errorf("Got portfolio metrics %v, want %v", got, want)
	}
}
