package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/alecthomas/kingpin/v2"
	"github.com/dwalker-sabiogroup/dependency-track-exporter/internal/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	sloghttp "github.com/samber/slog-http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	envAddress string = "DEPENDENCY_TRACK_ADDR"
	envAPIKey  string = "DEPENDENCY_TRACK_API_KEY"
)

func init() {
	prometheus.MustRegister(collectors.NewBuildInfoCollector())
}

func main() {
	var (
		profilingConfig = kingpin.Flag("web.pprof-listen-address", "Address to listen on for pprof").Default(":9917").String()
		webConfig       = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry").Default(":9916").String()
		metricsPath     = kingpin.Flag("web.metrics-path", "Path under which to expose metrics").Default("/metrics").String()
		dtAddress       = kingpin.Flag("dtrack.address", fmt.Sprintf("Dependency-Track server address (can also be set with $%s)", envAddress)).Default("http://localhost:8080").Envar(envAddress).String()
		dtAPIKey        = kingpin.Flag("dtrack.api-key", fmt.Sprintf("Dependency-Track API key (can also be set with $%s)", envAPIKey)).Envar(envAPIKey).Required().String()
		promlogConfig   = promlog.Config{}
	)

	flag.AddFlags(kingpin.CommandLine, &promlogConfig)
	kingpin.Version(version.Print("dependency_track_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	config := sloghttp.Config{
		WithSpanID:  true,
		WithTraceID: true,
	}

	logger.Info(fmt.Sprintf("Starting depnendency_track_exporter %s", version.Info()))
	logger.Info("Build context " + version.BuildContext())

	c, err := dtrack.NewClient(*dtAddress, dtrack.WithAPIKey(*dtAPIKey))
	if err != nil {
		logger.Error("Error creating client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	e := exporter.New(c, logger)

	prometheus.MustRegister(e)

	mux := http.NewServeMux()

	handler := sloghttp.Recovery(mux)
	handler = sloghttp.NewWithConfig(logger, config)(handler)

	mux.Handle(*metricsPath, otelhttp.WithRouteTag(*metricsPath, promhttp.Handler()))

	mux.Handle("/", otelhttp.WithRouteTag("/", http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})))

	srvc := make(chan struct{})
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Fatal(http.ListenAndServe(*profilingConfig, http.DefaultServeMux))
	}()

	go func() {
		err := http.ListenAndServe(*webConfig,
			otelhttp.NewHandler(
				handler,
				"server",
				otelhttp.WithMessageEvents(
					otelhttp.ReadEvents,
					otelhttp.WriteEvents,
				),
			),
		)

		if err != nil {
			logger.Error(err.Error())
		}

	}()

	for {
		select {
		case <-term:
			logger.Info("Received SIGTERM, exiting gracefully...")
			os.Exit(0)
		case <-srvc:
			os.Exit(1)
		}
	}
}
