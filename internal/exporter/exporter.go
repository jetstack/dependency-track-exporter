package exporter

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/jetstack/dependency-track-exporter/internal/dependencytrack"
)

const (
	// Namespace is the metrics namespace of the exporter
	Namespace string = "dependency_track"
)

// Exporter exports metrics from a Dependency-Track server
type Exporter struct {
	Client *dependencytrack.Client
	Logger log.Logger
}

// HandlerFunc handles requests to /metrics
func (e *Exporter) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registry := prometheus.NewRegistry()

		if err := e.collectPortfolioMetrics(registry); err != nil {
			level.Error(e.Logger).Log("err", err)
			http.Error(w, fmt.Sprintf("error: %s", err), http.StatusInternalServerError)
			return
		}

		if err := e.collectProjectMetrics(registry); err != nil {
			level.Error(e.Logger).Log("err", err)
			http.Error(w, fmt.Sprintf("error: %s", err), http.StatusInternalServerError)
			return
		}

		// Serve
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

func (e *Exporter) collectPortfolioMetrics(registry *prometheus.Registry) error {
	var (
		inheritedRiskScore = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(Namespace, "portfolio", "inherited_risk_score"),
				Help: "The inherited risk score of the whole portfolio.",
			},
		)
		vulnerabilities = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(Namespace, "portfolio", "vulnerabilities"),
				Help: "Number of vulnerabilities across the whole portfolio, by severity.",
			},
			[]string{
				"severity",
			},
		)
		findings = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(Namespace, "portfolio", "findings"),
				Help: "Number of findings across the whole portfolio, audited and unaudited.",
			},
			[]string{
				"audited",
			},
		)
	)
	registry.MustRegister(
		inheritedRiskScore,
		vulnerabilities,
		findings,
	)

	portfolioMetrics, err := e.Client.GetCurrentPortfolioMetrics()
	if err != nil {
		return err
	}

	inheritedRiskScore.Set(portfolioMetrics.InheritedRiskScore)

	severities := map[string]int32{
		"CRITICAL":   portfolioMetrics.Critical,
		"HIGH":       portfolioMetrics.High,
		"MEDIUM":     portfolioMetrics.Medium,
		"LOW":        portfolioMetrics.Low,
		"UNASSIGNED": portfolioMetrics.Unassigned,
	}
	for severity, v := range severities {
		vulnerabilities.With(prometheus.Labels{
			"severity": severity,
		}).Set(float64(v))
	}

	findingsAudited := map[string]int32{
		"true":  portfolioMetrics.FindingsAudited,
		"false": portfolioMetrics.FindingsUnaudited,
	}
	for audited, v := range findingsAudited {
		findings.With(prometheus.Labels{
			"audited": audited,
		}).Set(float64(v))
	}

	return nil
}

func (e *Exporter) collectProjectMetrics(registry *prometheus.Registry) error {
	var (
		info = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(Namespace, "project", "info"),
				Help: "Project information.",
			},
			[]string{
				"uuid",
				"name",
				"version",
				"classifier",
				"active",
				"tags",
			},
		)
		vulnerabilities = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(Namespace, "project", "vulnerabilities"),
				Help: "Number of vulnerabilities for a project by severity.",
			},
			[]string{
				"uuid",
				"name",
				"version",
				"severity",
			},
		)
		policyViolations = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(Namespace, "project", "policy_violations"),
				Help: "Policy violations for a project.",
			},
			[]string{
				"uuid",
				"name",
				"version",
				"type",
				"state",
				"analysis",
				"suppressed",
			},
		)
		lastBOMImport = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(Namespace, "project", "last_bom_import"),
				Help: "Last BOM import date, represented as a Unix timestamp.",
			},
			[]string{
				"uuid",
				"name",
				"version",
			},
		)
		inheritedRiskScore = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: prometheus.BuildFQName(Namespace, "project", "inherited_risk_score"),
				Help: "Inherited risk score for a project.",
			},
			[]string{
				"uuid",
				"name",
				"version",
			},
		)
	)
	registry.MustRegister(
		info,
		vulnerabilities,
		policyViolations,
		lastBOMImport,
		inheritedRiskScore,
	)

	projects, err := e.Client.GetProjects()
	if err != nil {
		return err
	}

	for _, project := range projects {
		projTags := ","
		for _, t := range project.Tags {
			projTags = projTags + t.Name + ","
		}
		info.With(prometheus.Labels{
			"uuid":       project.UUID,
			"name":       project.Name,
			"version":    project.Version,
			"classifier": project.Classifier,
			"active":     strconv.FormatBool(project.Active),
			"tags":       projTags,
		}).Set(1)

		severities := map[string]int32{
			"CRITICAL":   project.Metrics.Critical,
			"HIGH":       project.Metrics.High,
			"MEDIUM":     project.Metrics.Medium,
			"LOW":        project.Metrics.Low,
			"UNASSIGNED": project.Metrics.Unassigned,
		}
		for severity, v := range severities {
			vulnerabilities.With(prometheus.Labels{
				"uuid":     project.UUID,
				"name":     project.Name,
				"version":  project.Version,
				"severity": severity,
			}).Set(float64(v))
		}
		lastBOMImport.With(prometheus.Labels{
			"uuid":    project.UUID,
			"name":    project.Name,
			"version": project.Version,
		}).Set(float64(project.LastBomImport.Unix()))

		inheritedRiskScore.With(prometheus.Labels{
			"uuid":    project.UUID,
			"name":    project.Name,
			"version": project.Version,
		}).Set(project.Metrics.InheritedRiskScore)

		// Initialize all the possible violation series with a 0 value so that it
		// properly records increments from 0 -> 1
		for _, possibleType := range dependencytrack.PolicyViolationTypes {
			for _, possibleState := range dependencytrack.PolicyViolationStates {
				// If there isn't any analysis for a policy
				// violation then the value in the UI is
				// actually empty. So let's represent that in
				// these metrics as a possible analysis state.
				for _, possibleAnalysis := range append(dependencytrack.ViolationAnalysisStates, "") {
					for _, possibleSuppressed := range []string{"true", "false"} {
						policyViolations.With(prometheus.Labels{
							"uuid":       project.UUID,
							"name":       project.Name,
							"version":    project.Version,
							"type":       possibleType,
							"state":      possibleState,
							"analysis":   possibleAnalysis,
							"suppressed": possibleSuppressed,
						})
					}
				}
			}
		}
	}

	violations, err := e.Client.GetViolations(true)
	if err != nil {
		return err
	}

	for _, violation := range violations {
		var (
			analysisState string
			suppressed    string = "false"
		)
		if analysis := violation.Analysis; analysis != nil {
			analysisState = analysis.AnalysisState
			suppressed = strconv.FormatBool(analysis.IsSuppressed)
		}
		policyViolations.With(prometheus.Labels{
			"uuid":       violation.Project.UUID,
			"name":       violation.Project.Name,
			"version":    violation.Project.Version,
			"type":       violation.Type,
			"state":      violation.PolicyCondition.Policy.ViolationState,
			"analysis":   analysisState,
			"suppressed": suppressed,
		}).Inc()
	}

	return nil
}
