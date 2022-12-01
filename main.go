package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/log/level"
	"github.com/jetstack/dependency-track-exporter/internal/exporter"
	"github.com/nscuro/dtrack-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	envAddress string = "DEPENDENCY_TRACK_ADDR"
	envAPIKey  string = "DEPENDENCY_TRACK_API_KEY"
)

func init() {
	prometheus.MustRegister(version.NewCollector(exporter.Namespace + "_exporter"))
}

func main() {
	var (
		webConfig     = webflag.AddFlags(kingpin.CommandLine, ":9916")
		metricsPath   = kingpin.Flag("web.metrics-path", "Path under which to expose metrics").Default("/metrics").String()
		dtAddress     = kingpin.Flag("dtrack.address", fmt.Sprintf("Dependency-Track server address (can also be set with $%s)", envAddress)).Default("http://localhost:8080").Envar(envAddress).String()
		dtAPIKey      = kingpin.Flag("dtrack.api-key", fmt.Sprintf("Dependency-Track API key (can also be set with $%s)", envAPIKey)).Envar(envAPIKey).Required().String()
		promlogConfig = promlog.Config{}
	)

	flag.AddFlags(kingpin.CommandLine, &promlogConfig)
	kingpin.Version(version.Print(exporter.Namespace + "_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(&promlogConfig)

	level.Info(logger).Log("msg", fmt.Sprintf("Starting %s_exporter %s", exporter.Namespace, version.Info()))
	level.Info(logger).Log("msg", fmt.Sprintf("Build context %s", version.BuildContext()))

	c, err := dtrack.NewClient(*dtAddress, dtrack.WithAPIKey(*dtAPIKey))
	if err != nil {
		level.Error(logger).Log("msg", "Error creating client", "err", err)
		os.Exit(1)
	}

	e := exporter.Exporter{
		Client: c,
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

	srvc := make(chan struct{})
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		srv := &http.Server{}
		if err := web.ListenAndServe(srv, webConfig, logger); err != http.ErrServerClosed {
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
