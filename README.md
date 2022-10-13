# StecaGrid 2000 Exporter
Prometheus exporter for StecaGrid Inverter 2000 written in Go.

Metrics are fetched from `http://<steca-grid-ip>/measurements.xml`


## Getting started

1. Run `make` to build the binary
2. Setup a service with systemd ([example](./systemd/stecagrid-exporter.service))

## Supported settings
```
λ stecagrid-exporter -h
Usage of stecagrid-exporter:
  -frequency int
    	Polling frequency in seconds (default 5)
  -steca-ip string
    	StecaGrid IP address (default "192.168.50.144")
  -steca-path string
    	StecaGrid path (default "/measurements.xml")
```

## Grafana dashboard

![](img/dashboard.png?raw=true)
