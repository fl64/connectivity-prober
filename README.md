# Connectivity Prober

A lightweight Go-based utility that periodically checks the availability of network targets (HTTP, ICMP) and exposes Prometheus-style metrics.

## Build

```bash
git clone https://github.com/fl64/connectivity-prober.git
cd connectivity-prober
go build -o connectivity-prober
```

## Usage

```bash
./connectivity-prober \
  --interval=5000 \
  --target.ping=ya.ru,8.8.8.8 \
  --target.http=http://example.com,https://google.com  \
  --metrics.port=9090 \
  --metrics.file=./metrics.prom \
  --log.level=info
```

## CLI flags

```bash
--interval xx                      # Check interval in milliseconds, default: 5000
--target.ping	target1,target2,...  # Comma-separated list of ICMP targets
--target.http	target1,target2,...  # Comma-separated list of HTTP targets
--metrics.port xxxx                # Port for Prometheus metrics server, default: 8080
--metrics.file path/to/file.prom   # File path to save metrics
--log.level level                  # Log level (debug"	info	warn	error), default: warn
```
