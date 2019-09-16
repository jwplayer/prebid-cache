package backends

import (
	"context"
	"fmt"
)

type MemoryBackend struct {
	db map[string]string
}

func (b *MemoryBackend) Get(ctx context.Context, key string) (string, error) {
	v, ok := b.db[key]
	if !ok {
		return "", fmt.Errorf("Not found")
	}

	return v, nil
}

func (b *MemoryBackend) MultiPut(ctx context.Context, payloads []Payload) error {
	for _, payload := range payloads {
		key := payload.Key
		value := payload.Value
		b.db[key] = value
	}
	
	return nil
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		db: make(map[string]string),
	}
}
