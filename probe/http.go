package probe

import (
	"context"
	"net/http"
	"time"

	"log/slog"

	"github.com/fl64/connectivity-prober/metrics"
)

type HTTPProbe struct {
	Metrics *metrics.Metrics
	Logger  *slog.Logger
}

func (h *HTTPProbe) Run(ctx context.Context, target string) {
	start := time.Now()

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
	if err != nil {
		h.Logger.Warn("HTTP request creation failed", "target", target, "error", err)
		return
	}

	h.Logger.Debug("Running HTTP request", "target", target)
	resp, err := client.Do(req)
	if err != nil {
		h.Logger.Warn("HTTP request failed", "target", target, "error", err)
		return
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200, 301, 302:
		// OK
	default:
		h.Logger.Warn("HTTP status not allowed", "target", target, "status", resp.StatusCode)
		return
	}

	duration := time.Since(start).Seconds()
	h.Metrics.SuccessCount.WithLabelValues("http", target).Inc()
	h.Metrics.LatencyHist.WithLabelValues("http", target).Observe(duration)
	h.Logger.Info("HTTP request succeeded", "target", target, "rtt", duration, "status", resp.StatusCode)
}
