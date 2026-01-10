package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nxo/engine/internal/config"
	"github.com/redis/go-redis/v9"
)

// Redis wraps the Redis client
type Redis struct {
	client *redis.Client
}

// New creates a new Redis cache instance
func New(cfg *config.RedisConfig) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &Redis{client: client}
}

// Get retrieves a value from cache
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// GetJSON retrieves and unmarshals a JSON value
func (r *Redis) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Set stores a value in cache
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// SetJSON marshals and stores a JSON value
func (r *Redis) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, expiration).Err()
}

// Delete removes a key from cache
func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}

// SetNX sets a value only if the key doesn't exist (for locks)
func (r *Redis) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

// Ping checks Redis connectivity
func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (r *Redis) Close() error {
	return r.client.Close()
}

// Client returns the underlying Redis client
func (r *Redis) Client() *redis.Client {
	return r.client
}

// --- Token Blacklist ---

const blacklistPrefix = "token:blacklist:"

// BlacklistToken adds a token to the blacklist
func (r *Redis) BlacklistToken(ctx context.Context, tokenID string, expiration time.Duration) error {
	return r.Set(ctx, blacklistPrefix+tokenID, "1", expiration)
}

// IsTokenBlacklisted checks if a token is blacklisted
func (r *Redis) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	return r.Exists(ctx, blacklistPrefix+tokenID)
}
