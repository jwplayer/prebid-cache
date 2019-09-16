package decorators_test

import (
	"context"
	"testing"

	"github.com/prebid/prebid-cache/backends"
	"github.com/prebid/prebid-cache/backends/decorators"
)

func TestExcessiveTTL(t *testing.T) {
	delegate := &ttlCapturer{}
	wrapped := decorators.LimitTTLs(delegate, 100)
	wrapped.MultiPut(context.Background(), []backends.Payload{backends.Payload{Key: "foo", Value: "bar", TtlSeconds: 200}})
	if delegate.lastTTL != 100 {
		t.Errorf("lastTTL should be %d. Got %d", 100, delegate.lastTTL)
	}
}

func TestSafeTTL(t *testing.T) {
	delegate := &ttlCapturer{}
	wrapped := decorators.LimitTTLs(delegate, 100)
	wrapped.MultiPut(context.Background(), []backends.Payload{backends.Payload{Key: "foo", Value: "bar", TtlSeconds: 50}})
	if delegate.lastTTL != 50 {
		t.Errorf("lastTTL should be %d. Got %d", 50, delegate.lastTTL)
	}
}

type ttlCapturer struct {
	lastTTL int
}

func (c *ttlCapturer) MultiPut(ctx context.Context, payloads []backends.Payload) error {
	c.lastTTL = payloads[0].TtlSeconds
	return nil
}

func (c *ttlCapturer) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}
