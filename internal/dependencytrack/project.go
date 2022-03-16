package dependencytrack

import (
	"net/http"
)

// Project is a project in Dependency-Track
type Project struct {
	Name          string         `json:"name"`
	Version       string         `json:"version"`
	Classifier    string         `json:"classifier"`
	Active        bool           `json:"active"`
	LastBomImport Time           `json:"lastBomImport"`
	Metrics       ProjectMetrics `json:"metrics"`
	Tags          []ProjectTag   `json:"tags"`
	UUID          string         `json:"uuid"`
}

// ProjectTag is a project's tag
type ProjectTag struct {
	Name string `json:"name"`
}

// ProjectMetrics are metrics for the project
type ProjectMetrics struct {
	Critical           int32   `json:"critical"`
	High               int32   `json:"high"`
	Low                int32   `json:"low"`
	Medium             int32   `json:"medium"`
	Unassigned         int32   `json:"unassigned"`
	InheritedRiskScore float64 `json:"inheritedRiskScore"`
}

// GetProjects returns a list of all projects
func (c *Client) GetProjects() ([]*Project, error) {
	req, err := c.newRequest(http.MethodGet, "/api/v1/project", nil)
	if err != nil {
		return nil, err
	}

	out := []*Project{}
	if err := c.do(req, &out); err != nil {
		return nil, err
	}

	return out, nil
}
