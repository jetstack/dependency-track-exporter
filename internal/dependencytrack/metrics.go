package dependencytrack

import (
	"net/http"
)

// PortfolioMetrics are metrics for the whole portfolio
type PortfolioMetrics struct {
	Critical           int32   `json:"critical"`
	High               int32   `json:"high"`
	Low                int32   `json:"low"`
	Medium             int32   `json:"medium"`
	Unassigned         int32   `json:"unassigned"`
	InheritedRiskScore float64 `json:"inheritedRiskScore"`
	FindingsAudited    int32   `json:"findingsAudited"`
	FindingsUnaudited  int32   `json:"findingsUnaudited"`
}

// GetCurrentPortfolioMetrics returns the current metrics for the whole
// portfolio
func (c *Client) GetCurrentPortfolioMetrics() (*PortfolioMetrics, error) {
	req, err := c.newRequest(http.MethodGet, "/api/v1/metrics/portfolio/current", nil, nil)
	if err != nil {
		return nil, err
	}

	out := &PortfolioMetrics{}
	if err := c.do(req, out); err != nil {
		return nil, err
	}

	return out, nil
}
