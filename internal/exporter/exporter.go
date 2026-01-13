package exporter

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

const (
	// Namespace is the metrics namespace of the exporter
	Namespace string = "dependency_track"
)

// Exporter exports metrics from a Dependency-Track server
type Exporter struct {
	Client                     *dtrack.Client
	Logger                     log.Logger
	ProjectTags                []string
	InitializeViolationMetrics bool

	mutex    sync.RWMutex
	registry *prometheus.Registry
}

// HandlerFunc handles requests to /metrics
func (e *Exporter) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		e.mutex.RLock()
		registry := e.registry
		e.mutex.RUnlock()

		if registry == nil {
			http.Error(w, "Exporter not yet initialized", http.StatusServiceUnavailable)
			return
		}

		// Serve
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

// Run starts the background polling of Dependency-Track metrics
func (e *Exporter) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	level.Info(e.Logger).Log("msg", "Starting background poller", "interval", interval)

	// Initial poll
	e.poll(ctx)

	for {
		select {
		case <-ctx.Done():
			level.Info(e.Logger).Log("msg", "Stopping background poller")
			return
		case <-ticker.C:
			e.poll(ctx)
		}
	}
}

func (e *Exporter) poll(ctx context.Context) {
	level.Debug(e.Logger).Log("msg", "Polling Dependency-Track metrics")
	registry := prometheus.NewRegistry()
	registry.MustRegister(version.NewCollector(Namespace + "_exporter"))

	if err := e.collectPortfolioMetrics(ctx, registry); err != nil {
		level.Error(e.Logger).Log("msg", "Error collecting portfolio metrics", "err", err)
	}

	if err := e.collectProjectMetrics(ctx, registry); err != nil {
		level.Error(e.Logger).Log("msg", "Error collecting project metrics", "err", err)
	}

	e.mutex.Lock()
	e.registry = registry
	e.mutex.Unlock()
	level.Debug(e.Logger).Log("msg", "Successfully updated metrics cache")
}

func (e *Exporter) collectPortfolioMetrics(ctx context.Context, registry *prometheus.Registry) error {
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

	portfolioMetrics, err := e.Client.Metrics.LatestPortfolioMetrics(ctx)
	if err != nil {
		return err
	}

	inheritedRiskScore.Set(portfolioMetrics.InheritedRiskScore)

	severities := map[string]int{
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

	findingsAudited := map[string]int{
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

func (e *Exporter) collectProjectMetrics(ctx context.Context, registry *prometheus.Registry) error {
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

	matchedProjects := make(map[string]struct{})

	err := e.forEachProject(ctx, func(project dtrack.Project) error {
		projectUUID := project.UUID.String()
		matchedProjects[projectUUID] = struct{}{}

		var tags []string
		for _, t := range project.Tags {
			tags = append(tags, t.Name)
		}

		info.WithLabelValues(
			projectUUID,
			project.Name,
			project.Version,
			project.Classifier,
			strconv.FormatBool(project.Active),
			strings.Join(tags, ","),
		).Set(1)

		severities := map[string]int{
			"CRITICAL":   project.Metrics.Critical,
			"HIGH":       project.Metrics.High,
			"MEDIUM":     project.Metrics.Medium,
			"LOW":        project.Metrics.Low,
			"UNASSIGNED": project.Metrics.Unassigned,
		}
		for severity, v := range severities {
			vulnerabilities.WithLabelValues(
				projectUUID,
				project.Name,
				project.Version,
				severity,
			).Set(float64(v))
		}
		lastBOMImport.WithLabelValues(
			projectUUID,
			project.Name,
			project.Version,
		).Set(float64(project.LastBOMImport))

		inheritedRiskScore.WithLabelValues(
			projectUUID,
			project.Name,
			project.Version,
		).Set(project.Metrics.InheritedRiskScore)

		// Initialize all the possible violation series with a 0 value so that it
		// properly records increments from 0 -> 1.
		// Note: This accounts for 72 series per project.
		if e.InitializeViolationMetrics {
			for _, possibleType := range []string{"LICENSE", "OPERATIONAL", "SECURITY"} {
				for _, possibleState := range []string{"INFO", "WARN", "FAIL"} {
					for _, possibleAnalysis := range []dtrack.ViolationAnalysisState{
						dtrack.ViolationAnalysisStateApproved,
						dtrack.ViolationAnalysisStateRejected,
						dtrack.ViolationAnalysisStateNotSet,
						"",
					} {
						for _, possibleSuppressed := range []string{"true", "false"} {
							policyViolations.WithLabelValues(
								projectUUID,
								project.Name,
								project.Version,
								possibleType,
								possibleState,
								string(possibleAnalysis),
								possibleSuppressed,
							).Set(0)
						}
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = e.forEachPolicyViolation(ctx, func(violation dtrack.PolicyViolation) error {
		if _, ok := matchedProjects[violation.Project.UUID.String()]; !ok {
			return nil
		}
		var (
			analysisState string
			suppressed    string = "false"
		)
		if analysis := violation.Analysis; analysis != nil {
			analysisState = string(analysis.State)
			suppressed = strconv.FormatBool(analysis.Suppressed)
		}
		policyViolations.WithLabelValues(
			violation.Project.UUID.String(),
			violation.Project.Name,
			violation.Project.Version,
			violation.Type,
			violation.PolicyCondition.Policy.ViolationState,
			analysisState,
			suppressed,
		).Inc()
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *Exporter) forEachProject(ctx context.Context, fn func(dtrack.Project) error) error {
	if len(e.ProjectTags) == 0 {
		return dtrack.ForEach(func(po dtrack.PageOptions) (dtrack.Page[dtrack.Project], error) {
			return e.Client.Project.GetAll(ctx, po)
		}, fn)
	}

	seen := make(map[string]struct{})
	for _, tag := range e.ProjectTags {
		err := dtrack.ForEach(func(po dtrack.PageOptions) (dtrack.Page[dtrack.Project], error) {
			return e.Client.Project.GetAllByTag(ctx, tag, po)
		}, func(p dtrack.Project) error {
			id := p.UUID.String()
			if _, ok := seen[id]; ok {
				return nil
			}
			seen[id] = struct{}{}
			return fn(p)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Exporter) forEachPolicyViolation(ctx context.Context, fn func(dtrack.PolicyViolation) error) error {
	return dtrack.ForEach(func(po dtrack.PageOptions) (dtrack.Page[dtrack.PolicyViolation], error) {
		return e.Client.PolicyViolation.GetAll(ctx, true, po)
	}, fn)
}

func (e *Exporter) fetchProjects(ctx context.Context) ([]dtrack.Project, error) {
	var projects []dtrack.Project
	err := e.forEachProject(ctx, func(p dtrack.Project) error {
		projects = append(projects, p)
		return nil
	})
	return projects, err
}

func (e *Exporter) fetchPolicyViolations(ctx context.Context) ([]dtrack.PolicyViolation, error) {
	var violations []dtrack.PolicyViolation
	err := e.forEachPolicyViolation(ctx, func(v dtrack.PolicyViolation) error {
		violations = append(violations, v)
		return nil
	})
	return violations, err
}
