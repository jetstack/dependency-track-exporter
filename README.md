# Dependency-Track Exporter

Exports Prometheus metrics for [Dependency-Track](https://dependencytrack.org/).

## Usage

```
usage: dependency-track-exporter [<flags>]

Flags:
  -h, --help                Show context-sensitive help (also try --help-long and --help-man).
      --web.config.file=""  [EXPERIMENTAL] Path to configuration file that can enable TLS or authentication.
      --web.listen-address=":9916"
                            Address to listen on for web interface and telemetry.
      --web.metrics-path="/metrics"
                            Path under which to expose metrics
      --dtrack.address=DTRACK.ADDRESS
                            Dependency-Track server address (default: http://localhost:8080 or $DEPENDENCY_TRACK_ADDR)
      --dtrack.api-key=DTRACK.API-KEY
                            Dependency-Track API key (default: $DEPENDENCY_TRACK_API_KEY)
      --dtrack.project-tags=DTRACK.PROJECT-TAGS
                            Comma-separated list of project tags to filter on
      --dtrack.poll-interval=6h
                            Interval to poll Dependency-Track for metrics
      --dtrack.initialize-violation-metrics
                            Initialize all possible violation metric combinations to 0 (default: true)
      --log.level=info      Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt   Output format of log messages. One of: [logfmt, json]
      --version             Show application version.
```

The API key the exporter uses needs to have the following permissions:
- `VIEW_POLICY_VIOLATION`
- `VIEW_PORTFOLIO`

## Metrics

| Metric                                          | Meaning                                                               | Labels                                           |
| ----------------------------------------------- | --------------------------------------------------------------------- | ------------------------------------------------ |
| dependency_track_portfolio_inherited_risk_score | The inherited risk score of the whole portfolio.                      |                                                        |
| dependency_track_portfolio_vulnerabilities      | Number of vulnerabilities across the whole portfolio, by severity.    | severity                                               |
| dependency_track_portfolio_findings             | Number of findings across the whole portfolio, audited and unaudited. | audited                                                |
| dependency_track_project_info                   | Project information.                                                  | uuid, name, version, classifier, active, tags          |
| dependency_track_project_vulnerabilities        | Number of vulnerabilities for a project by severity.                  | uuid, name, version, severity                          |
| dependency_track_project_policy_violations      | Policy violations for a project.                                      | uuid, name, version, type, state, analysis, suppressed |
| dependency_track_project_last_bom_import        | Last BOM import date, represented as a Unix timestamp.                | uuid, name, version                                    |
| dependency_track_project_inherited_risk_score   | Inherited risk score for a project.                                   | uuid, name, version                                    |

## Performance & Memory Optimization

If you have a very large Dependency-Track portfolio, the exporter can consume significant memory during polling due to the high cardinality of policy violation metrics.

### High-Cardinality Metrics
By default, the exporter initializes 72 unique metric series for every project (combinations of violation types, states, etc.) to ensure they record `0` instead of being absent. 

To significantly reduce memory usage, you can disable this behavior:

```bash
--dtrack.initialize-violation-metrics=false
```

When disabled, metric series will only be created when an actual violation is detected.

### Streaming
The exporter uses streaming pagination to fetch data from Dependency-Track, ensuring that memory usage remains stable even as your portfolio grows.

## Example queries

Retrieve the number of `WARN` policy violations that have not been analyzed or
suppressed:

```
dependency_track_project_policy_violations{state="WARN",analysis!="APPROVED",analysis!="REJECTED",suppressed="false"} > 0
```

Exclude inactive projects:

```
dependency_track_project_policy_violations{state="WARN",analysis!="APPROVED",analysis!="REJECTED",suppressed="false"} > 0
and on(uuid) dependency_track_project_info{active="true"}
```

Only include projects tagged with `prod`:

```
dependency_track_project_policy_violations{state="WARN",analysis!="APPROVED",analysis!="REJECTED",suppressed="false"} > 0
and on(uuid) dependency_track_project_info{active="true",tags=~".*,prod,.*"}
```

Or, join the tags label into the returned series. Filtering on active/tag could
then happen in alert routes:

```
(dependency_track_project_policy_violations{state="WARN",analysis!="APPROVED",analysis!="REJECTED",suppressed="false"} > 0)
* on (uuid) group_left(tags,active) dependency_track_project_info
```
