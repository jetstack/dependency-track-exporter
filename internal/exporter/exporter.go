package exporter

import (
	"context"
	"log/slog"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	client             *dtrack.Client
	logger             *slog.Logger
	vulnerabilities    *prometheus.Desc
	policyViolations   *prometheus.Desc
	lastBOMImport      *prometheus.Desc
	inheritedRiskScore *prometheus.Desc
}

func New(cl *dtrack.Client, l *slog.Logger) *Exporter {
	return &Exporter{
		client: cl,
		logger: l,
		vulnerabilities: prometheus.NewDesc(
			"dependency_track_project_vulnerability",
			"Number of vulnerabilities for a project by severity",
			[]string{"name", "uuid", "version", "severity"},
			nil,
		),
		policyViolations: prometheus.NewDesc(
			"dependency_track_project_policy_violation",
			"Policy violations for a project",
			[]string{"name", "uuid", "version", "state"},
			nil,
		),
		lastBOMImport: prometheus.NewDesc(
			"dependency_track_project_last_bom_import",
			"Last BOM import date, represented as a Unix timestamp",
			[]string{"name", "uuid", "version"},
			nil,
		),
		inheritedRiskScore: prometheus.NewDesc(
			"dependency_track_project_inherited_risk_score",
			"Inherited risk score for a project",
			[]string{"name", "uuid", "version"},
			nil,
		),
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	projects, err := e.fetchProjects(ctx)
	if err != nil {
		e.logger.WarnContext(ctx, err.Error())
	}

	for _, p := range e.filterProjects(projects) {
		ch <- prometheus.MustNewConstMetric(e.vulnerabilities, prometheus.GaugeValue, float64(p.Metrics.Critical), p.Name, p.UUID.String(), p.Version, "critical")
		ch <- prometheus.MustNewConstMetric(e.vulnerabilities, prometheus.GaugeValue, float64(p.Metrics.High), p.Name, p.UUID.String(), p.Version, "high")
		ch <- prometheus.MustNewConstMetric(e.vulnerabilities, prometheus.GaugeValue, float64(p.Metrics.Medium), p.Name, p.UUID.String(), p.Version, "medium")
		ch <- prometheus.MustNewConstMetric(e.vulnerabilities, prometheus.GaugeValue, float64(p.Metrics.Critical), p.Name, p.UUID.String(), p.Version, "low")
		ch <- prometheus.MustNewConstMetric(e.vulnerabilities, prometheus.GaugeValue, float64(p.Metrics.Unassigned), p.Name, p.UUID.String(), p.Version, "unassigned")

		ch <- prometheus.MustNewConstMetric(e.policyViolations, prometheus.GaugeValue, float64(p.Metrics.PolicyViolationsFail), p.Name, p.UUID.String(), p.Version, "fail")
		ch <- prometheus.MustNewConstMetric(e.policyViolations, prometheus.GaugeValue, float64(p.Metrics.PolicyViolationsWarn), p.Name, p.UUID.String(), p.Version, "warn")
		ch <- prometheus.MustNewConstMetric(e.policyViolations, prometheus.GaugeValue, float64(p.Metrics.PolicyViolationsInfo), p.Name, p.UUID.String(), p.Version, "info")
		ch <- prometheus.MustNewConstMetric(e.policyViolations, prometheus.GaugeValue, float64(p.Metrics.PolicyViolationsUnaudited), p.Name, p.UUID.String(), p.Version, "unaudited")

		ch <- prometheus.MustNewConstMetric(e.lastBOMImport, prometheus.GaugeValue, float64(p.LastBOMImport), p.Name, p.UUID.String(), p.Version)
		ch <- prometheus.MustNewConstMetric(e.inheritedRiskScore, prometheus.GaugeValue, p.Metrics.InheritedRiskScore, p.Name, p.UUID.String(), p.Version)
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(e, ch)
}

func (e *Exporter) fetchProjects(ctx context.Context) ([]dtrack.Project, error) {
	return dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.Project], error) {
		return e.client.Project.GetAll(ctx, po)
	})
}

func (e *Exporter) filterProjects(projects []dtrack.Project) []dtrack.Project {
	m := make(map[string]dtrack.Project, len(projects))

	for _, p := range projects {
		val, ok := m[p.Name]
		if !ok {
			m[p.Name] = p
		}

		if val.Metrics.FirstOccurrence < p.Metrics.FirstOccurrence {
			m[p.Name] = p
		}
	}

	filtered := make([]dtrack.Project, 0, len(projects))

	for _, p := range m {
		filtered = append(filtered, p)
	}

	return filtered
}
