package dependencytrack

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

var (
	// PolicyViolationStates are the possible states for a violation
	PolicyViolationStates = []string{
		"INFO",
		"WARN",
		"FAIL",
	}

	// PolicyViolationTypes are the possible types for a violation
	PolicyViolationTypes = []string{
		"LICENSE",
		"OPERATIONAL",
		"SECURITY",
	}

	// ViolationAnalysisStates are the possible states for a violation
	// analysis
	ViolationAnalysisStates = []string{
		"APPROVED",
		"REJECTED",
		"NOT_SET",
	}
)

// PolicyViolation is a violation
type PolicyViolation struct {
	Analysis        *ViolationAnalysis `json:"analysis,omitempty"`
	PolicyCondition PolicyCondition    `json:"policyCondition"`
	Project         Project            `json:"project"`
	Type            string             `json:"type"`
}

// PolicyCondition contains the policy
type PolicyCondition struct {
	Policy Policy `json:"policy"`
}

// Policy is a policy
type Policy struct {
	ViolationState string `json:"violationState,omitempty"`
}

// ViolationAnalysis is the analysis decisions that have been made for the
// violation
type ViolationAnalysis struct {
	AnalysisState string `json:"analysisState"`
	IsSuppressed  bool   `json:"isSuppressed,omitempty"`
}

// GetAllViolations returns violations for the entire portfolio. Suppressed
// violations are omitted unless suppressed is true
func (c *Client) GetAllViolations(suppressed bool) ([]*PolicyViolation, error) {
	// dependency track per default only returns a 100 items per get, therefore we need to iterate over all PolicyViolation pages to get all PolicyViolation

	// all PolicyViolation found in pagination
	allPolicyViolations := []*PolicyViolation{}
	// the last project found in the last pagination page result
	lastPaginationPage := 1
	// state var to show if all PolicyViolation projects where found
	foundAll := false

	for !foundAll {
		req, err := c.GetViolations(suppressed, lastPaginationPage, 100)
		if err != nil {
			return nil, err
		}
		if len(req) == 0 {
			foundAll = true
			break
		}
		allPolicyViolations = append(allPolicyViolations, req...)
		lastPaginationPage++
	}
	return allPolicyViolations, nil
}

// GetViolations returns violations for the entire portfolio. Suppressed
// violations are omitted unless suppressed is true
func (c *Client) GetViolations(suppressed bool, pageNumber int, pageSize int) ([]*PolicyViolation, error) {
	params := url.Values{}
	params.Add("suppressed", strconv.FormatBool(suppressed))
	req, err := c.newRequest(http.MethodGet, fmt.Sprintf("/api/v1/violation?%s&pageSize=%d&pageNumber=%d", params.Encode(), pageSize, pageNumber), nil)
	if err != nil {
		return nil, err
	}

	out := []*PolicyViolation{}
	if err := c.do(req, &out); err != nil {
		return nil, err
	}

	return out, nil
}
