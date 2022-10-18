package dependencytrack

import (
	"fmt"
	"net/http"
	"strconv"
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
	// dependency track per default only returns a 100 items per get, therefore we need to iterate over all pages to get all projects

	// all project found trough pagination
	allProjectsFound := []*Project{}
	// the last project found in the last pagination page result
	lastFoundProject := Project{}
	lastProjectPage := 1
	// state var to show if all projects where found
	didFindAllProjects := false

	for !didFindAllProjects {
		fmt.Printf("lastProjectPage: %v\n", lastProjectPage)
		req, err := c.GetProjects(lastProjectPage, 100)
		if err != nil {
			return nil, err
		}
		fmt.Println(req[len(req)-1].UUID)

		if req[len(req)-1].UUID == lastFoundProject.UUID {
			didFindAllProjects = true
			break
		}

		allProjectsFound = append(allProjectsFound, req...)
		lastFoundProject = req[len(req)-1] // TODO fix
	}
	return allProjectsFound, nil
}

// GetProjects returns a list of all projects with pagination
func (c *Client) GetProjects(pageNumber int, pageSize int) ([]*Project, error) {
	var headers = map[string]string{}
	headers["pageNumber"] = strconv.Itoa(pageNumber)
	headers["pageSize"] = strconv.Itoa(pageSize)
	req, err := c.newRequest(http.MethodGet, "/api/v1/project", headers, nil)
	if err != nil {
		return nil, err
	}

	out := []*Project{}
	if err := c.do(req, &out); err != nil {
		return nil, err
	}

	return out, nil
}
