# Rate Limiter

  Built this project to understand how rate limiting works in backend systems and to get hands on experience with concurrency in Go. 

  ## Features
  
  - Token bucket rate limiting
  - Separate bucket for each client
  - Thread safe implementation
  - Request statistics tracking
  - HTTP API endpoints
  - Middleware support for protecting routes
  - Automatic cleanup of inactive buckets
  - Unit tests

  ## Limitations
  
- In-memory only - doesn't share state across multiple instances
- retryAfter is approximate (1/refillRate)
- No per-route or per-user configuration

## Running

```bash
go run main.go
```

## Testing

Run unit tests:

```bash
go test ./...
```

Run race detection (tested in WSL):

```bash
go test -race ./...
```
