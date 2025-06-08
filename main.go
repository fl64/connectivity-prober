package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fl64/connectivity-prober/metrics"
	"github.com/fl64/connectivity-prober/probe"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
)

var (
	interval    = flag.Int("interval", 5000, "Interval in milliseconds between checks")
	metricsPort = flag.Int("metrics.port", 8080, "Port for Prometheus metrics server")
	logLevel    = flag.String("log.level", "warn", "Log level: debug, info, warn, error")
	pingTargets = flag.String("target.ping", "", "Comma-separated list of ping targets")
	httpTargets = flag.String("target.http", "", "Comma-separated list of HTTP targets")
	metricsFile = flag.String("metrics.file", "", "File to write metrics in Prometheus format (optional)")
)

func writeMetricsToFile(path string, logger *slog.Logger) {
	// Собираем все метрики
	metrics, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		logger.Error("Failed to gather metrics", "error", err)
		return
	}

	tmpFile, err := os.CreateTemp("", "metrics-*.prom")
	if err != nil {
		logger.Error("Failed to create temp file", "error", err)
		return
	}
	defer os.Remove(tmpFile.Name())

	encoder := expfmt.NewEncoder(tmpFile, expfmt.FmtText)
	for _, mf := range metrics {
		if err := encoder.Encode(mf); err != nil {
			logger.Error("Failed to encode metrics", "error", err)
			return
		}
	}

	if err := os.Rename(tmpFile.Name(), path); err != nil {
		logger.Error("Failed to replace metrics file", "error", err)
	}
}

func startMetricsFileWriter(path string, interval time.Duration, logger *slog.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			writeMetricsToFile(path, logger)
		}
	}
}

func main() {
	flag.Parse()

	var level slog.Level
	if err := level.UnmarshalText([]byte(*logLevel)); err != nil {
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(log.Default().Writer(), &slog.HandlerOptions{Level: level}))

	pingTargetsList := parseTargets(*pingTargets)
	httpTargetsList := parseTargets(*httpTargets)

	metricRegistry := metrics.NewMetrics()
	prometheus.MustRegister(metricRegistry.SuccessCount, metricRegistry.LatencyHist)

	pingChecker := &probe.PingProbe{
		Metrics: metricRegistry,
		Logger:  logger,
	}
	httpChecker := &probe.HTTPProbe{
		Metrics: metricRegistry,
		Logger:  logger,
	}

	go func() {
		metricsAddr := fmt.Sprintf(":%d", *metricsPort)
		logger.Warn("Starting metrics server", "addr", metricsAddr)
		logger.Error("Metrics server exited", "err", http.ListenAndServe(metricsAddr, promhttp.Handler()))
	}()

	if *metricsFile != "" {
		go startMetricsFileWriter(*metricsFile, 10*time.Second, logger)
		logger.Info("Metrics file writer started", "path", *metricsFile, "interval", "10s")
	}

	ticker := time.NewTicker(time.Duration(*interval) * time.Millisecond)
	logger.Info("Starting prober loop")
	for {
		<-ticker.C
		logger.Debug("Running probes...")
		runProbe(pingChecker, pingTargetsList)
		runProbe(httpChecker, httpTargetsList)
	}
}

func parseTargets(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

func runProbe(probeType probe.Probe, targets []string) {
	for _, target := range targets {
		go func(p probe.Probe, t string) {
			p.Run(context.Background(), t)
		}(probeType, target)
	}
}
