package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/kaizendorks/terraform-cloud-exporter/internal/collector"
	"github.com/kaizendorks/terraform-cloud-exporter/internal/setup"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Build information. Populated at build-time via ldflags.
var (
	Version   string
	Commit    string
	GoVersion = runtime.Version()
	BuildDate string
)

func newHandler(metrics collector.Metrics, config setup.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Use request context for cancellation when connection gets closed.
		ctx := r.Context()
		// If a timeout is configured via the Prometheus header, add it to the context.
		if v := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
			timeoutSeconds, err := strconv.ParseFloat(v, 64)
			if err != nil {
				config.Logger.Error("Failed to parse timeout from Prometheus header", "err", err)
			} else {
				// Create new timeout context with request context as parent.
				ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds*float64(time.Second)))
				defer cancel()
				// Overwrite request with timeout context.
				r = r.WithContext(ctx)
			}
		}

		registry := prometheus.NewRegistry()
		registry.MustRegister(collector.New(ctx, config, metrics))

		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			registry,
		}
		// Delegate http serving to Prometheus client library, which will call collector.Collect.
		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{EnableOpenMetrics: true})
		h.ServeHTTP(w, r)
	}
}

func main() {
	config := setup.NewConfig()
	config.Logger.Info("Starting tf_exporter", "version", Version, "revision", Commit)
	config.Logger.Log(context.Background(), slog.LevelDebug, "Build Context", "go", GoVersion, "date", BuildDate)

	handlerFunc := newHandler(collector.NewMetrics(), config)
	http.Handle("/metrics", promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, handlerFunc))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Terraform Cloud/Enterprise Exporter</title></head>
			<body>
			<h1>Terraform Cloud/Enterprise Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`))
	})

	srv := &http.Server{Addr: config.ListenAddress}

	go func() {
		config.Logger.Info("Listening on address", "address", config.ListenAddress)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			config.Logger.Error("Error starting HTTP server", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	config.Logger.Info("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		config.Logger.Error("Server forced to shutdown", "err", err)
	}
}
