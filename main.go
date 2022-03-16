package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"github.com/ribbybibby/dependency-track-exporter/internal/dependencytrack"
	"github.com/ribbybibby/dependency-track-exporter/internal/exporter"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	prometheus.MustRegister(version.NewCollector(exporter.Namespace + "_exporter"))
}

func main() {
	var (
		webConfig     = webflag.AddFlags(kingpin.CommandLine)
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9916").String()
		metricsPath   = kingpin.Flag("web.metrics-path", "Path under which to expose metrics").Default("/metrics").String()
		dtAddress     = kingpin.Flag("dtrack.address", fmt.Sprintf("Dependency-Track server address (default: %s or $%s)", dependencytrack.DefaultAddress, dependencytrack.EnvAddress)).String()
		dtAPIKey      = kingpin.Flag("dtrack.api-key", fmt.Sprintf("Dependency-Track API key (default: $%s)", dependencytrack.EnvAPIKey)).String()
		promlogConfig = promlog.Config{}
	)

	flag.AddFlags(kingpin.CommandLine, &promlogConfig)
	kingpin.Version(version.Print(exporter.Namespace + "_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(&promlogConfig)

	level.Info(logger).Log("msg", fmt.Sprintf("Starting %s_exporter %s", exporter.Namespace, version.Info()))
	level.Info(logger).Log("msg", fmt.Sprintf("Build context %s", version.BuildContext()))

	var opts []dependencytrack.Option
	if *dtAddress != "" {
		opts = append(opts, dependencytrack.WithAddress(*dtAddress))
	}
	if *dtAPIKey != "" {
		opts = append(opts, dependencytrack.WithAPIKey(*dtAPIKey))
	}
	e := exporter.Exporter{
		Client: dependencytrack.New(opts...),
		Logger: logger,
	}

	http.HandleFunc(*metricsPath, e.HandlerFunc())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
						 <head><title>Dependency-Track Exporter</title></head>
						 <body>
						 <h1>Dependency-Track Exporter</h1>
						 <p><a href='` + *metricsPath + `'>Metrics</a></p>
						 </body>
						 </html>`))
	})

	srv := &http.Server{Addr: *listenAddress}
	srvc := make(chan struct{})
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
		if err := web.ListenAndServe(srv, *webConfig, logger); err != http.ErrServerClosed {
			level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
			close(srvc)
		}
	}()

	for {
		select {
		case <-term:
			level.Info(logger).Log("msg", "Received SIGTERM, exiting gracefully...")
			os.Exit(0)
		case <-srvc:
			os.Exit(1)
		}
	}
}
