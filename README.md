# Dependency Track Exporter

Exports Prometheus metrics for [Dependency Track](https://dependencytrack.org/).

## TODO

- Tests for the client.
- Github Actions workflows for tests and release

## Usage

```
usage: dependency-track-exporter [<flags>]

Flags:
  -h, --help               Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9219"
                           Address to listen on for web interface and telemetry.
      --web.metrics-path="/metrics"
                           Path under which to expose metrics
      --dtrack.address=DTRACK.ADDRESS
                           Dependency Track server address (default: http://localhost:8080 or $DEPENDENCY_TRACK_ADDR)
      --dtrack.api-key=DTRACK.API-KEY
                           Dependency Track API key (default: $DEPENDENCY_TRACK_API_KEY)
      --log.level=info     Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt  Output format of log messages. One of: [logfmt, json]
      --version            Show application version.
```

## Metrics

| Metric                                          | Meaning                                                | Labels                                           |
| ----------------------------------------------- | ------------------------------------------------------ | ------------------------------------------------ |
| dependency_track_portfolio_inherited_risk_score | The inherited risk score of the whole portfolio.       |                                                  |
| dependency_track_project_active                 | Is this project active?                                | uuid, name, version                              |
| dependency_track_project_vulnerabilities        | Number of vulnerabilities for a project by severity.   | uuid, name, version, severity                    |
| dependency_track_project_policy_violations      | Policy violations for a project.                       | uuid, name, version, state, analysis, suppressed |
| dependency_track_project_last_bom_import        | Last BOM import date, represented as a Unix timestamp. | uuid, name, version                              |
| dependency_track_project_inherited_risk_score   | Inherited risk score for a project.                    | uuid, name, version                              |

## Example queries

Retrieve the number of `WARN` policy violations that have not been analyzed or
suppressed:

```
dependency_track_project_policy_violations{state="WARN",analysis!="APPROVED",analysis!="REJECTED",suppressed="false"}
```

Exclude inactive projects:

```
sum(dependency_track_project_policy_violations{state="WARN",analysis!="APPROVED",analysis!="REJECTED",suppressed="false"}) by (uuid,name,version) 
* on(uuid,name,version) dependency_track_project_active
```
