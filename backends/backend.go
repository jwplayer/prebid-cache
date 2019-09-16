package backends

import (
	"context"
)

// Backend interface for storing data
type Backend interface {
	Get(ctx context.Context, key string) (string, error)
	MultiPut(ctx context.Context, payloads []Payload) error
}

// Payload struct for storing the data to store in the backend
type Payload struct {
	Key string
	Value string
	TtlSeconds int
}
