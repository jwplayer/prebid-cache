package backends

import (
	"context"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/prebid/prebid-cache/config"
)

// MemcacheConfig is used to configure the cluster
type MemcacheConfig struct {
	hosts []string
}

// Memcache Object use to implement backend interface
type Memcache struct {
	client *memcache.Client
}

// NewMemcacheBackend create a new memcache backend
func NewMemcacheBackend(cfg config.Memcache) *Memcache {
	c := &Memcache{}
	mc := memcache.New(cfg.Hosts...)
	c.client = mc
	return c
}

func (mc *Memcache) Get(ctx context.Context, key string) (string, error) {
	res, err := mc.client.Get(key)

	if err != nil {
		return "", err
	}

	return string(res.Value), nil
}

func (mc *Memcache) MultiPut(ctx context.Context, payloads []Payload) error {
	for _, payload := range payloads {
		key := payload.Key
		value := payload.Value
		ttlSeconds := payload.TtlSeconds

		err := mc.client.Set(&memcache.Item{
			Expiration: int32(ttlSeconds),
			Key:        key,
			Value:      []byte(value),
		})

		if err != nil {
			return err
		}
	}
	
	return nil
}
