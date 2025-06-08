package probe

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/fl64/connectivity-prober/metrics"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type PingProbe struct {
	Metrics *metrics.Metrics
	Logger  *slog.Logger
}

func (p *PingProbe) Run(ctx context.Context, target string) {
	p.Logger.Debug("Running ping", "target", target)

	dst, err := net.ResolveIPAddr("ip", target)
	if err != nil {
		p.Logger.Warn("Failed to resolve IP address", "target", target, "error", err)
		return
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		p.Logger.Warn("Failed to listen ICMP socket", "target", target, "error", err)
		return
	}
	defer conn.Close()

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   12345,
			Seq:  1,
			Data: []byte("HELLO-R-U-THERE"),
		},
	}

	wb, _ := msg.Marshal(nil)
	start := time.Now()
	if _, err := conn.WriteTo(wb, dst); err != nil {
		p.Logger.Warn("Failed to send ICMP packet", "target", target, "error", err)
		return
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	rb := make([]byte, 1500)
	n, peer, err := conn.ReadFrom(rb)
	if err != nil {
		p.Logger.Warn("Failed to receive ICMP response", "target", target, "error", err)
		return
	}

	reply, err := icmp.ParseMessage(1, rb[:n])
	if err != nil {
		p.Logger.Warn("Failed to parse ICMP reply", "target", target, "error", err)
		return
	}

	switch reply.Type {
	case ipv4.ICMPTypeEchoReply:
		duration := time.Since(start).Seconds()
		p.Logger.Info("Ping succeeded", "target", target, "rtt", duration, "from", peer)
		p.Metrics.SuccessCount.WithLabelValues("ping", target).Inc()
		p.Metrics.LatencyHist.WithLabelValues("ping", target).Observe(duration)
	default:
		p.Logger.Warn("Unexpected ICMP response", "target", target, "type", reply.Type)
	}
}
