package decorators

import (
	"context"

	"github.com/prebid/prebid-cache/backends"
)

// LimitTTLs wraps the delegate and makes sure that it never gets TTLs which exceed the max.
func LimitTTLs(delegate backends.Backend, maxTTLSeconds int) backends.Backend {
	return ttlLimited{
		Backend:       delegate,
		maxTTLSeconds: maxTTLSeconds,
	}
}

type ttlLimited struct {
	backends.Backend
	maxTTLSeconds int
}

func (l ttlLimited) MultiPut(ctx context.Context, payloads []backends.Payload) error {
	limitedPayloads := []backends.Payload{}
	for _, payload := range payloads {
		if l.maxTTLSeconds < payload.TtlSeconds {
			limitedPayloads = append(limitedPayloads, backends.Payload{Key: payload.Key, Value: payload.Value, TtlSeconds: l.maxTTLSeconds})
		} else {
			limitedPayloads = append(limitedPayloads, payload)
		}
	}
	return l.Backend.MultiPut(ctx, limitedPayloads)
}
