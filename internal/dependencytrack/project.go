package dependencytrack

import (
	"fmt"
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

// GetAllProjects returns a list of all projects with the help of pagination
func (c *Client) GetAllProjects() ([]*Project, error) {
	// dependency track per default only returns a 100 items per get, therefore we need to iterate over allProjects pages to get allProjects projects

	// allProjects found in pagination
	allProjects := []*Project{}
	// the last project found in the last pagination page result
	lastPaginationPage := 1
	// state var to show if allProjects projects where found
	foundAll := false

	for !foundAll {
		req, err := c.GetProjects(lastPaginationPage, 100)
		if err != nil {
			return nil, err
		}
		if len(req) == 0 {
			foundAll = true
			break
		}
		allProjects = append(allProjects, req...)
		lastPaginationPage++
	}
	return allProjects, nil
}

// GetProjects returns a list of all projects with pagination
func (c *Client) GetProjects(pageNumber int, pageSize int) ([]*Project, error) {
	req, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/project?pageSize=%d&pageNumber=%d", pageSize, pageNumber), nil)
	if err != nil {
		return nil, err
	}

	out := []*Project{}
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out, nil
}
