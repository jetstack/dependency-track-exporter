package dependencytrack

import (
	"net/http"
)

// PortfolioMetrics are metrics for the whole portfolio
type PortfolioMetrics struct {
	InheritedRiskScore float64 `json:"inheritedRiskScore"`
}

// GetCurrentPortfolioMetrics returns the current metrics for the whole
// portfolio
func (c *Client) GetCurrentPortfolioMetrics() (*PortfolioMetrics, error) {

	req, err := c.newRequest(http.MethodGet, "/api/v1/metrics/portfolio/current", nil)
	if err != nil {
		return nil, err
	}

	out := &PortfolioMetrics{}
	if err := c.do(req, out); err != nil {
		return nil, err
	}

	return out, nil
}
