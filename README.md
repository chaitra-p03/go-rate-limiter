# Rate Limiter

A concurrency-safe HTTP rate limiter implemented in Go using the Token Bucket algorithm.  
The server supports per-client rate limiting, automatic token refilling, request statistics and cleanup of inactive clients.

## Features

- Token Bucket based rate limiting
- Per-client request isolation using individual buckets
- Lazy token refill based on elapsed time
- Concurrency-safe request handling
- Automatic cleanup of inactive buckets
- Middleware support for HTTP endpoints
- Request statistics tracking
- Unit tests with race detection

## How It Works

Each client is assigned an independent token bucket.

A bucket contains:
- Maximum token capacity
- Current available tokens
- Token refill rate
- Last refill timestamp

Every request consumes one token. If tokens are available, the request is allowed. Otherwise, it is rejected with HTTP `429 Too Many Requests`.

The limiter uses lazy refill instead of a background refill thread. Tokens are recalculated only when a bucket is accessed by checking how much time has passed since the last refill.

This avoids unnecessary work for inactive clients.


## Architecture

```
HTTP Request
      |
      v
Rate Limiter Middleware
      |
      v
RateLimiterManager
      |
      v
map[clientID]*Bucket
      |
      v
Token Bucket
```

## Concurrency Design

The rate limiter uses two levels of synchronization:

- `RateLimiterManager` mutex protects the bucket map during creation and deletion.
- Individual bucket mutexes protect token updates and refill calculations.

This allows different clients to update their own buckets concurrently.

Global request counters use atomic operations for safe concurrent updates.

## API Endpoints

### Check Request

`POST /check`

Request:

```json
{
  "identifier": "user1",
  "capacity": 10,
  "refill_rate": 1
}
```

Response:

```json
{
  "allowed": true,
  "remaining": 8,
  "limit": 10
}
```

If the limit is exceeded:

```json
{
  "allowed": false,
  "remaining": 0,
  "limit": 10,
  "retry_after": 1
}
```

Returns:

```
429 Too Many Requests
```

---

### Statistics

`GET /stats`

Returns:

```json
{
  "total": 1000,
  "allowed": 950,
  "denied": 50,
  "rejection_rate": "5.00%",
  "active_clients": 20
}
```

## Running the Server

```bash
go run main.go
```

Server starts on:

```
localhost:8080
```

## Running Tests

```bash
go test ./...
```

Run with race detector:

```bash
go test -race ./...
```


## Benchmarking

Load testing was performed using `hey` to evaluate concurrent request handling.

Create a request body:

```json
{
  "identifier": "user1",
  "capacity": 100000,
  "refill_rate": 1000
}
```

Run benchmark:

```bash
hey -m POST \
-H "Content-Type: application/json" \
-D body.json \
-n 10000 \
-c 100 \
http://localhost:8080/check
```

Test configuration:

- 10,000 total requests
- 100 concurrent clients
- 25,000+ requests/sec throughput

The benchmark verifies concurrent request handling, token bucket updates and HTTP response correctness under load.