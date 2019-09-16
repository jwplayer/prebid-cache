package decorators

import (
	"context"
	"testing"

	"github.com/prebid/prebid-cache/backends"
)

func TestLargePayload(t *testing.T) {
	delegate := &successfulBackend{}
	wrapped := EnforceSizeLimit(delegate, 5)
	
	assertBadPayloadError(t, wrapped.MultiPut(context.Background(), []backends.Payload{backends.Payload{Key: "foo", Value: "123456", TtlSeconds: 0}}))
}

func TestAcceptablePayload(t *testing.T) {
	delegate := &successfulBackend{}
	wrapped := EnforceSizeLimit(delegate, 5)
	assertNilError(t, wrapped.MultiPut(context.Background(), []backends.Payload{backends.Payload{Key: "foo", Value: "12345", TtlSeconds: 0}}))
}

func assertBadPayloadError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Errorf("Expected an error, but got none.")
	}
	if _, ok := err.(*BadPayloadSize); !ok {
		t.Errorf("Expected a BadPayloadSize error. Got %#v", err)
	}
}

func assertNilError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
}

type successfulBackend struct{}

func (b *successfulBackend) Get(ctx context.Context, key string) (string, error) {
	return "some-value", nil
}

func (b *successfulBackend) MultiPut(ctx context.Context, payloads []backends.Payload) error {
	return nil
}
