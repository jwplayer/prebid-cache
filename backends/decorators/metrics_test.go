package decorators

import (
	"context"
	"fmt"
	"testing"

	"github.com/prebid/prebid-cache/backends"
	"github.com/prebid/prebid-cache/config"
	"github.com/prebid/prebid-cache/metrics"
	"github.com/prebid/prebid-cache/metrics/metricstest"
)

type failedBackend struct{}

func (b *failedBackend) Get(ctx context.Context, key string) (string, error) {
	return "", fmt.Errorf("Failure")
}

func (b *failedBackend) Put(ctx context.Context, key string, value string, ttlSeconds int) error {
	return fmt.Errorf("Failure")
}

func TestGetSuccessMetrics(t *testing.T) {
	m := metrics.CreateMetrics(config.NewConfig().Metrics)
	rawBackend := backends.NewMemoryBackend()
	rawBackend.Put(context.Background(), "foo", "xml<vast></vast>", 0)
	backend := LogMetrics(rawBackend, m)
	backend.Get(context.Background(), "foo")

	metricstest.AssertSuccessMetricsExist(t, m.GetsBackend)
}

func TestGetErrorMetrics(t *testing.T) {
	m := metrics.CreateMetrics(config.NewConfig().Metrics)
	backend := LogMetrics(&failedBackend{}, m)
	backend.Get(context.Background(), "foo")

	metricstest.AssertErrorMetricsExist(t, m.GetsBackend)
}

func TestPutSuccessMetrics(t *testing.T) {
	m := metrics.CreateMetrics(config.NewConfig().Metrics)
	backend := LogMetrics(backends.NewMemoryBackend(), m)
	backend.Put(context.Background(), "foo", "xml<vast></vast>", 0)

	assertSuccessMetricsExist(t, m.PutsBackend)
	if m.PutsBackend.XmlRequest.Count() != 1 {
		t.Errorf("An xml request should have been logged.")
	}
	if m.PutsBackend.DefinesTTL.Count() != 0 {
		t.Errorf("An event for TTL defined shouldn't be logged if the TTL was 0")
	}
}

func TestTTLDefinedMetrics(t *testing.T) {
	m := metrics.CreateMetrics(config.NewConfig().Metrics)
	backend := LogMetrics(backends.NewMemoryBackend(), m)
	backend.Put(context.Background(), "foo", "xml<vast></vast>", 1)
	if m.PutsBackend.DefinesTTL.Count() != 1 {
		t.Errorf("An event for TTL defined should be logged if the TTL is not 0")
	}
}

func TestPutErrorMetrics(t *testing.T) {
	m := metrics.CreateMetrics(config.NewConfig().Metrics)
	backend := LogMetrics(&failedBackend{}, m)
	backend.Put(context.Background(), "foo", "xml<vast></vast>", 0)

	assertErrorMetricsExist(t, m.PutsBackend)
	if m.PutsBackend.XmlRequest.Count() != 1 {
		t.Errorf("The request should have been counted.")
	}
}

func TestJsonPayloadMetrics(t *testing.T) {
	m := metrics.CreateMetrics(config.NewConfig().Metrics)
	backend := LogMetrics(backends.NewMemoryBackend(), m)
	backend.Put(context.Background(), "foo", "json{\"key\":\"value\"", 0)
	backend.Get(context.Background(), "foo")

	if m.PutsBackend.JsonRequest.Count() != 1 {
		t.Errorf("A json Put should have been logged.")
	}
}

func TestPutSizeSampling(t *testing.T) {
	m := metrics.CreateMetrics(config.NewConfig().Metrics)
	payload := `json{"key":"value"}`
	backend := LogMetrics(backends.NewMemoryBackend(), m)
	backend.Put(context.Background(), "foo", payload, 0)

	if m.PutsBackend.RequestLength.Count() != 1 {
		t.Errorf("A request size sample should have been logged.")
	}
}

func TestInvalidPayloadMetrics(t *testing.T) {
	m := metrics.CreateMetrics(config.NewConfig().Metrics)
	backend := LogMetrics(backends.NewMemoryBackend(), m)
	backend.Put(context.Background(), "foo", "bar", 0)
	backend.Get(context.Background(), "foo")

	if m.PutsBackend.InvalidRequest.Count() != 1 {
		t.Errorf("A Put request of invalid format should have been logged.")
	}
}

func assertSuccessMetricsExist(t *testing.T, entry *metrics.MetricsEntryByFormat) {
	t.Helper()
	if entry.Duration.Count() != 1 {
		t.Errorf("The request duration should have been counted.")
	}
	if entry.BadRequest.Count() != 0 {
		t.Errorf("No Bad requests should have been counted.")
	}
	if entry.Errors.Count() != 0 {
		t.Errorf("No Errors should have been counted.")
	}
}

func assertErrorMetricsExist(t *testing.T, entry *metrics.MetricsEntryByFormat) {
	t.Helper()
	if entry.Duration.Count() != 0 {
		t.Errorf("The request duration should not have been counted.")
	}
	if entry.BadRequest.Count() != 0 {
		t.Errorf("No Bad requests should have been counted.")
	}
	if entry.Errors.Count() != 1 {
		t.Errorf("An Error should have been counted.")
	}
}
