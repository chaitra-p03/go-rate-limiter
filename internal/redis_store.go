package internal

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const allowScriptSource = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local idle_ttl_ms = tonumber(ARGV[3])

local redis_time = redis.call("TIME")
local now_ms = tonumber(redis_time[1]) * 1000
	+ math.floor(tonumber(redis_time[2]) / 1000)

local values = redis.call(
	"HMGET", key,
	"tokens",
	"last_refill_ms"
)

local tokens = tonumber(values[1])
local last_refill_ms = tonumber(values[2])

if tokens == nil then
	tokens = capacity
	last_refill_ms = now_ms
end

local elapsed_seconds = (now_ms - last_refill_ms) / 1000
if elapsed_seconds < 0 then
	elapsed_seconds = 0
end

tokens = math.min(capacity, tokens + elapsed_seconds * refill_rate)

local allowed = 0
if tokens >= 1 then
	tokens = tokens - 1
	allowed = 1
end

redis.call(
	"HSET", key,
	"tokens", tostring(tokens),
	"last_refill_ms", tostring(now_ms),
	"capacity", tostring(capacity),
	"refill_rate", tostring(refill_rate)
)

if idle_ttl_ms > 0 then
	redis.call("PEXPIRE", key, idle_ttl_ms)
end

return {allowed, tostring(tokens)}
`

var allowScript = redis.NewScript(allowScriptSource)

// RedisStore holds bucket state in Redis, shared by every server using it.
type RedisStore struct {
	client  redis.UniversalClient
	prefix  string
	idleTTL time.Duration
}

func NewRedisStore(
	client redis.UniversalClient,
	idleTTL time.Duration,
) *RedisStore {
	return &RedisStore{
		client:  client,
		prefix:  "rate-limit:",
		idleTTL: idleTTL,
	}
}

// Allow refills a client's bucket, then atomically checks and takes one token.
func (s *RedisStore) Allow(
	ctx context.Context,
	identifier string,
	capacity, refillRate float64,
) (bool, float64, error) {
	key := s.prefix + base64.RawURLEncoding.EncodeToString([]byte(identifier))

	result, err := allowScript.Run(
		ctx,
		s.client,
		[]string{key},
		capacity,
		refillRate,
		s.idleTTL.Milliseconds(),
	).Slice()
	if err != nil {
		return false, 0, fmt.Errorf("run rate-limit script: %w", err)
	}
	if len(result) != 2 {
		return false, 0, fmt.Errorf("unexpected rate-limit script result: %v", result)
	}

	allowedValue, err := strconv.ParseInt(fmt.Sprint(result[0]), 10, 64)
	if err != nil {
		return false, 0, fmt.Errorf("parse allowed result: %w", err)
	}

	remaining, err := strconv.ParseFloat(fmt.Sprint(result[1]), 64)
	if err != nil {
		return false, 0, fmt.Errorf("parse remaining tokens: %w", err)
	}

	return allowedValue == 1, remaining, nil
}