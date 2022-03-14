package dependencytrack

import (
	"encoding/json"
	"net/http"
	"time"
)

// Project is a project in Dependency Track
type Project struct {
	Name          string         `json:"name"`
	Version       string         `json:"version"`
	Active        bool           `json:"active"`
	LastBomImport time.Time      `json:"lastBomImport"`
	Metrics       ProjectMetrics `json:"metrics"`
	UUID          string         `json:"uuid"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *Project) UnmarshalJSON(data []byte) error {
	aux := struct {
		Name    string         `json:"name"`
		Version string         `json:"version"`
		Metrics ProjectMetrics `json:"metrics"`
		UUID    string         `json:"uuid"`

		// LastBomImport is a unix timestamp in the response but we want
		// to convert it to a time.Time
		LastBomImport int64 `json:"lastBomImport"`
	}{}
	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	p.Name = aux.Name
	p.Version = aux.Version
	p.Metrics = aux.Metrics
	p.UUID = aux.UUID
	p.LastBomImport = time.Unix(aux.LastBomImport, 0)

	return nil
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
