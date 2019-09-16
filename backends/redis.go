package backends

import (
	"context"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-redis/redis"
	"github.com/prebid/prebid-cache/config"
)

type Redis struct {
	cfg    config.Redis
	client *redis.Client
}

func NewRedisBackend(cfg config.Redis) *Redis {
	constr := cfg.Host + ":" + strconv.Itoa(cfg.Port)
	client := redis.NewClient(&redis.Options{
		Addr:     constr,
		Password: cfg.Password,
		DB:       cfg.Db,
	})

	_, err := client.Ping().Result()

	if err != nil {
		log.Fatalf("Error creating Redis backend: %v", err)
	}

	log.Infof("Connected to Redis at %s:%d", cfg.Host, cfg.Port)

	return &Redis{
		cfg:    cfg,
		client: client,
	}
}

func (redis *Redis) Get(ctx context.Context, key string) (string, error) {
	res, err := redis.client.Get(key).Result()

	if err != nil {
		return "", err
	}

	return string(res), nil
}

func (redis *Redis) MultiPut(ctx context.Context, payloads []Payload) error {
	pipeline := redis.client.Pipeline()
	for _, payload := range payloads {
		ttlSeconds := payload.TtlSeconds
		if ttlSeconds == 0 {
			ttlSeconds = redis.cfg.Expiration * 60
		}
		pipeline.Set(payload.Key, payload.Value, time.Duration(ttlSeconds)*time.Second)
	}
	_, err := pipeline.Exec()

	if err != nil {
		return err
	}

	return nil
}
