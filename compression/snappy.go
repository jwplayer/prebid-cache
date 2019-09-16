package compression

import (
	"context"

	"github.com/golang/snappy"
	"github.com/prebid/prebid-cache/backends"
)

// SnappyCompress runs snappy compression on data before saving it in the backend.
// For more info, see https://en.wikipedia.org/wiki/Snappy_(compression)
func SnappyCompress(backend backends.Backend) backends.Backend {
	return &snappyCompressor{
		delegate: backend,
	}
}

type snappyCompressor struct {
	delegate backends.Backend
}

func (s *snappyCompressor) MultiPut(ctx context.Context, payloads []backends.Payload) error {
	compressedPayloads := []backends.Payload{}
	for _, payload := range payloads {
		compressedPayloads = append(compressedPayloads, backends.Payload{
			Key: payload.Key,
			Value: string(snappy.Encode(nil, []byte(payload.Value))),
			TtlSeconds: payload.TtlSeconds})
	}
	return s.delegate.MultiPut(ctx, compressedPayloads)
}

func (s *snappyCompressor) Get(ctx context.Context, key string) (string, error) {
	compressed, err := s.delegate.Get(ctx, key)
	if err != nil {
		return "", err
	}

	decompressed, err := snappy.Decode(nil, []byte(compressed))
	if err != nil {
		return "", err
	}

	return string(decompressed), nil
}
