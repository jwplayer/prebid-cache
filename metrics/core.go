package metrics

import (
	"fmt"
	"log"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/prebid/prebid-cache/config"
	"github.com/rcrowley/go-metrics"
	datadog "github.com/syntaqx/go-metrics-datadog"
	influxdb "github.com/vrischmann/go-metrics-influxdb"
)

type MetricsEntry struct {
	Duration   metrics.Timer
	Errors     metrics.Meter
	BadRequest metrics.Meter
	Request    metrics.Meter
}

type MetricsEntryByFormat struct {
	Duration       metrics.Timer
	Errors         metrics.Meter
	BadRequest     metrics.Meter
	JsonRequest    metrics.Meter
	XmlRequest     metrics.Meter
	DefinesTTL     metrics.Meter
	InvalidRequest metrics.Meter
	RequestLength  metrics.Histogram
}

type ConnectionMetrics struct {
	ActiveConnections      metrics.Counter
	ConnectionCloseErrors  metrics.Meter
	ConnectionAcceptErrors metrics.Meter
}

func NewMetricsEntry(name string, r metrics.Registry) *MetricsEntry {
	return &MetricsEntry{
		Duration:   metrics.GetOrRegisterTimer(fmt.Sprintf("%s.request_duration", name), r),
		Errors:     metrics.GetOrRegisterMeter(fmt.Sprintf("%s.error_count", name), r),
		BadRequest: metrics.GetOrRegisterMeter(fmt.Sprintf("%s.bad_request_count", name), r),
		Request:    metrics.GetOrRegisterMeter(fmt.Sprintf("%s.request_count", name), r),
	}
}

func NewMetricsEntryBackendPuts(name string, r metrics.Registry) *MetricsEntryByFormat {
	return &MetricsEntryByFormat{
		Duration:       metrics.GetOrRegisterTimer(fmt.Sprintf("%s.request_duration", name), r),
		Errors:         metrics.GetOrRegisterMeter(fmt.Sprintf("%s.error_count", name), r),
		BadRequest:     metrics.GetOrRegisterMeter(fmt.Sprintf("%s.bad_request_count", name), r),
		JsonRequest:    metrics.GetOrRegisterMeter(fmt.Sprintf("%s.json_request_count", name), r),
		XmlRequest:     metrics.GetOrRegisterMeter(fmt.Sprintf("%s.xml_request_count", name), r),
		DefinesTTL:     metrics.GetOrRegisterMeter(fmt.Sprintf("%s.defines_ttl", name), r),
		InvalidRequest: metrics.GetOrRegisterMeter(fmt.Sprintf("%s.unknown_request_count", name), r),
		RequestLength:  metrics.GetOrRegisterHistogram(name+".request_size_bytes", r, metrics.NewExpDecaySample(1028, 0.015)),
	}
}

func NewConnectionMetrics(r metrics.Registry) *ConnectionMetrics {
	return &ConnectionMetrics{
		ActiveConnections:      metrics.GetOrRegisterCounter("connections.active_incoming", r),
		ConnectionAcceptErrors: metrics.GetOrRegisterMeter("connections.accept_errors", r),
		ConnectionCloseErrors:  metrics.GetOrRegisterMeter("connections.close_errors", r),
	}
}

type Metrics struct {
	Registry        metrics.Registry
	Puts            *MetricsEntry
	Gets            *MetricsEntry
	PutsBackend     *MetricsEntryByFormat
	GetsBackend     *MetricsEntry
	Connections     *ConnectionMetrics
	ExtraTTLSeconds metrics.Histogram
}

// Export begins sending metrics to the configured database.
// This method blocks indefinitely, so it should probably be run in a goroutine.
func (m *Metrics) Export(cfg config.Metrics) {
	switch cfg.Type {
	case config.MetricsDatadog:
		logrus.Infof("Metrics will be exported to Datadog with host=%s, port=%s", cfg.Datadog.Host, cfg.Datadog.Port)
		reporter, err := datadog.NewReporter(
			m.Registry,
			fmt.Sprintf("%s:%s", cfg.Datadog.Host, cfg.Datadog.Port),
			time.Second*cfg.Interval,
		)
		if err != nil {
			log.Fatal(err)
		}
		reporter.Client.Tags = cfg.Tags
		reporter.Flush()
	case config.MetricsInflux:
		logrus.Infof("Metrics will be exported to Influx with host=%s, db=%s, username=%s", cfg.Influx.Host, cfg.Influx.Database, cfg.Influx.Username)
		influxdb.InfluxDB(
			m.Registry,               // metrics registry
			time.Second*cfg.Interval, // interval
			cfg.Influx.Host,          // the InfluxDB url
			cfg.Influx.Database,      // your InfluxDB database
			cfg.Influx.Username,      // your InfluxDB user
			cfg.Influx.Password,      // your InfluxDB password
		)
	case config.MetricsNone:
		return
	default:
		logrus.Fatalf("Unrecognized config metrics.type: %s", cfg.Type)
	}
	return
}

func CreateMetrics(cfg config.Metrics) *Metrics {
	flushTime := time.Second * cfg.Interval
	r := metrics.NewPrefixedRegistry(cfg.Prefix)
	m := &Metrics{
		Registry:        r,
		Puts:            NewMetricsEntry("puts.current_url", r),
		Gets:            NewMetricsEntry("gets.current_url", r),
		PutsBackend:     NewMetricsEntryBackendPuts("puts.backend", r),
		GetsBackend:     NewMetricsEntry("gets.backend", r),
		Connections:     NewConnectionMetrics(r),
		ExtraTTLSeconds: metrics.GetOrRegisterHistogram("extra_ttl_seconds", r, metrics.NewUniformSample(5000)),
	}

	metrics.RegisterDebugGCStats(m.Registry)
	metrics.RegisterRuntimeMemStats(m.Registry)

	go metrics.CaptureRuntimeMemStats(m.Registry, flushTime)
	go metrics.CaptureDebugGCStats(m.Registry, flushTime)

	return m
}
