package backends

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prebid/prebid-cache/config"
	"github.com/stretchr/testify/assert"
)

// TestNewRedisBackend_ClusterMode tests that NewRedisBackend creates a cluster client when addrs are provided
// This test requires a running Redis cluster. It will be skipped if no cluster is available.
func TestNewRedisBackend_ClusterMode(t *testing.T) {
	// Check if cluster is available before running the test
	if !isRedisClusterAvailable(t) {
		t.Skip("Skipping cluster test: Redis cluster not available. Start cluster with: docker compose -f docker-compose.test-redis.yaml up -d (or docker-compose for V1)")
		return
	}

	cfg := config.Redis{
		Addrs:    []string{"127.0.0.1:7010", "127.0.0.1:7011", "127.0.0.1:7012"},
		Password: "",
		PoolSize: 10,
	}

	ctx := context.Background()
	
	// This will create a real cluster connection
	backend := NewRedisBackend(cfg, ctx)
	
	// Verify backend was created
	assert.NotNil(t, backend)
	assert.NotNil(t, backend.client)
	
	// Verify it's using cluster client (type assertion)
	_, isCluster := backend.client.(RedisClusterDBClient)
	assert.True(t, isCluster, "Backend should use RedisClusterDBClient in cluster mode")
}

// isRedisClusterAvailable checks if a Redis cluster is available for testing
func isRedisClusterAvailable(t *testing.T) bool {
	// Try to connect to one of the cluster nodes
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	testClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:7010",
	})
	defer testClient.Close()

	err := testClient.Ping(ctx).Err()
	return err == nil
}

// TestNewRedisBackend_SingleInstanceMode tests that NewRedisBackend creates a single instance client
// This test requires a running Redis instance. It will be skipped if no Redis is available.
func TestNewRedisBackend_SingleInstanceMode(t *testing.T) {
	// Check if Redis is available before running the test
	if !isRedisAvailable(t) {
		t.Skip("Skipping single instance test: Redis not available. Start Redis with: docker run -d -p 6380:6380 redis:7-alpine redis-server --port 6380")
		return
	}

	cfg := config.Redis{
		Host:     "127.0.0.1",
		Port:     6380,
		Password: "",
		Db:       0,
	}

	ctx := context.Background()
	
	backend := NewRedisBackend(cfg, ctx)
	
	// Verify backend was created
	assert.NotNil(t, backend)
	assert.NotNil(t, backend.client)
	
	// Verify it's using single instance client (type assertion)
	_, isSingle := backend.client.(RedisDBClient)
	assert.True(t, isSingle, "Backend should use RedisDBClient in single instance mode")
}

// isRedisAvailable checks if a Redis instance is available for testing
func isRedisAvailable(t *testing.T) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	testClient := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6380",
	})
	defer testClient.Close()

	err := testClient.Ping(ctx).Err()
	return err == nil
}

// TestRedisClusterDBClient_Get tests the cluster client Get method
func TestRedisClusterDBClient_Get(t *testing.T) {
	// This is a unit test using a mock - actual cluster testing requires integration test
	clusterBackend := &RedisBackend{}
	
	type testInput struct {
		redisClient RedisDB
		key         string
	}

	testCases := []struct {
		desc     string
		in       testInput
		expected string
		hasError bool
	}{
		{
			desc: "Cluster client Get() with existing key",
			in: testInput{
				redisClient: FakeRedisClient{
					Success:    true,
					StoredData: map[string]string{"test-key": "test-value"},
				},
				key: "test-key",
			},
			expected: "test-value",
			hasError: false,
		},
		{
			desc: "Cluster client Get() with non-existent key",
			in: testInput{
				redisClient: FakeRedisClient{
					Success:    false,
					StoredData: map[string]string{},
					ServerError: redis.Nil,
				},
				key: "non-existent",
			},
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range testCases {
		clusterBackend.client = tt.in.redisClient

		actualValue, actualErr := clusterBackend.Get(context.Background(), tt.in.key)

		if tt.hasError {
			assert.Error(t, actualErr, tt.desc)
		} else {
			assert.NoError(t, actualErr, tt.desc)
			assert.Equal(t, tt.expected, actualValue, tt.desc)
		}
	}
}

// TestRedisClusterDBClient_Put tests the cluster client Put method
func TestRedisClusterDBClient_Put(t *testing.T) {
	clusterBackend := &RedisBackend{}

	type testInput struct {
		redisClient  RedisDB
		key          string
		valueToStore string
		ttl          int
	}

	testCases := []struct {
		desc     string
		in       testInput
		hasError bool
	}{
		{
			desc: "Cluster client Put() successful",
			in: testInput{
				redisClient: FakeRedisClient{
					StoredData:  map[string]string{},
					Success:     true,
					ServerError: redis.Nil,
				},
				key:          "new-key",
				valueToStore: "new-value",
				ttl:          60,
			},
			hasError: false,
		},
		{
			desc: "Cluster client Put() with existing key (SetNX returns false)",
			in: testInput{
				redisClient: FakeRedisClient{
					StoredData:  map[string]string{"existing-key": "old-value"},
					Success:     false,
					ServerError: redis.Nil,
				},
				key:          "existing-key",
				valueToStore: "new-value",
				ttl:          60,
			},
			hasError: true, // Should return RECORD_EXISTS error
		},
	}

	for _, tt := range testCases {
		clusterBackend.client = tt.in.redisClient

		actualErr := clusterBackend.Put(context.Background(), tt.in.key, tt.in.valueToStore, tt.in.ttl)

		if tt.hasError {
			assert.Error(t, actualErr, tt.desc)
		} else {
			assert.NoError(t, actualErr, tt.desc)
		}
	}
}

