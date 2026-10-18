# Rate Limiter Design Evolution

## 001 - Simplifying client identity with a resolver

### Before
The middleware identified a client by IP, while later request handling also used an identifier from the JSON body. That meant a request could be associated with two different identities.

### Change
ResolveClient now chooses one key for the limiter:

- If X-API-Key is present, use it
- Otherwise, use the request IP

### Why
The limiter now has one consistent client identity. A two-level policy could be useful if the service explicitly wanted both per-IP and per-account limits; this project chose the simpler single-identity policy.

## 002 - Handling requests in middleware and returning rate-limit headers

The middleware calls `Allow`, which checks and consumes one token as a single operation. If allowed, it sets the rate-limit headers and calls the wrapped handler with `next.ServeHTTP(w, r)`. If denied, it returns HTTP 429 without calling the handler.

Responses include:

- `RateLimit-Limit`: configured bucket capacity
- `RateLimit-Remaining`: tokens remaining after the attempt
- `Retry-After`: estimated wait, sent when the request is denied

## 003 - Limiting the intended route and removing `/check`

The middleware is attached to `/api/data`. `/stats` is registered separately, so it remains accessible when a client has used its quota.

The debug `/check` endpoint was removed because it accepted client-supplied identifiers and rate-limit settings for manual testing. Automated tests now verify the limiter’s behavior.

## 004 - Separating bucket storage

The manager now uses the `Store` interface, with `MemoryStore` as its current implementation. `MemoryStore` still keeps the client buckets in an in-memory map.

This separates the manager from the map implementation. It does not add Redis or share state across server processes yet