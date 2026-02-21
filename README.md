[![codecov](https://codecov.io/gh/AnukritiSharma1609/caspage/branch/main/graph/badge.svg)](https://codecov.io/gh/AnukritiSharma1609/caspage)
[![Go Reference](https://pkg.go.dev/badge/github.com/AnukritiSharma1609/caspage.svg)](https://pkg.go.dev/github.com/AnukritiSharma1609/caspage)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/AnukritiSharma1609/caspage)](https://goreportcard.com/report/github.com/AnukritiSharma1609/caspage)

# caspage

`caspage` is a **lightweight, high-performance pagination library** for Apache Cassandra in Go, built on top of [`gocql`](https://github.com/gocql/gocql).

It provides **stateful and stateless pagination**, dynamic query filters, context handling, metrics hooks, and structured logging — all without changing your underlying Cassandra schema or application logic.

**Think of it as `gocql` pagination done right: simpler, safer, and production-ready.**

---

## Table of Contents

- [Why caspage?](#-why-caspage)
- [Features](#-features)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Usage Examples](#-usage-examples)
- [API Reference](#-api-reference)
- [Configuration](#-configuration)
- [Production Features](#-production-features)
- [Examples](#-examples)
- [Contributing](#-contributing)
- [License](#-license)

---

## Why caspage?

Pagination in Cassandra using `gocql` has always been tricky:

| Problem with raw gocql                              | How caspage solves it                                             |
|-----------------------------------------------------|-------------------------------------------------------------------|
| No direct API for cursor-based pagination           | Provides both stateful (`Next`) and stateless (`NextWithToken`) pagination |
| Page state tokens aren't REST-safe                  | Encodes and decodes them into portable Base64 tokens             |
| Requires manual handling of iterators               | Automatically manages page tokens and iterator lifecycle         |
| No previous page or cache support                   | Built-in token cache with `Previous()` navigation                |
| No built-in metrics, logging, or filters            | Ships with Prometheus hooks, structured logging, and query filters|
| No context awareness                                | Supports `context.Context` for cancellation and timeouts         |
| Complex filter handling                             | Dynamic `WHERE` clause building with operators (`>`, `<`, `IN`)  |

---

## Features

- **Simple API** – Paginate results using just `Next()` or `NextWithToken()`
- **Stateless pagination** – Tokens can be safely shared via REST APIs
- **Bidirectional navigation** – Move forward and backward between pages
- **Dynamic filters** – Add `WHERE` clauses with operators (`=`, `>`, `<`, `>=`, `<=`, `IN`)
- **Token cache** – Keep track of visited tokens in memory for backward navigation
- **Context-aware queries** – Use `context.Context` for safe cancellations and timeouts
- **Metrics hooks** – Plug in Prometheus (or any custom collector) easily
- **Structured logging** – Log query performance and pagination details
- **Column selection** – Fetch only specific columns to reduce payload size
- **Production-ready** – Thread-safe, tested, and optimized for high throughput
- **Drop-in compatible** – Works with existing `gocql` code, no schema changes needed

---

## Installation

```bash
go get github.com/AnukritiSharma1609/caspage
```

**Requirements:**
- Go 1.18 or higher
- `gocql` driver (automatically installed as dependency)

---

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/gocql/gocql"
    "github.com/AnukritiSharma1609/caspage/core"
)

func main() {
    // 1️⃣ Connect to Cassandra
    cluster := gocql.NewCluster("127.0.0.1")
    cluster.Keyspace = "my_keyspace"
    session, err := cluster.CreateSession()
    if err != nil {
        log.Fatal(err)
    }
    defer session.Close()

    // 2️⃣ Create a paginator
    paginator := core.NewPaginator(
        &core.RealSession{Session: session},
        "SELECT * FROM users",
        core.Options{PageSize: 100},
    )

    // 3️⃣ Fetch first page
    results, nextToken, err := paginator.Next()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Fetched %d rows\n", len(results))
    fmt.Printf("Next token: %s\n", nextToken)

    // 4️⃣ Fetch next page using token
    if nextToken != "" {
        results2, nextToken2, err := paginator.NextWithToken(nextToken)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Fetched %d more rows\n", len(results2))
        fmt.Printf("Next token: %s\n", nextToken2)
    }
}
```

---

## Usage Examples

### Basic Pagination

```go
// Create paginator
p := core.NewPaginator(
    &core.RealSession{Session: session},
    "SELECT * FROM products",
    core.Options{PageSize: 50},
)

// Fetch pages
results, token, _ := p.Next()          // First page
results2, token2, _ := p.NextWithToken(token)  // Second page
```

### Pagination with Filters

```go
// Apply dynamic filters
p := core.NewPaginator(
    &core.RealSession{Session: session},
    "SELECT * FROM orders",
    core.Options{
        PageSize: 100,
        Filters: map[string]interface{}{
            "user_id":    "12345",          // WHERE user_id = ?
            "amount >":   1000,              // AND amount > ?
            "status IN":  []string{"pending", "approved"}, // AND status IN (?, ?)
        },
    },
)

results, token, _ := p.Next()
```

**Supported filter operators:**
- `=` (default)
- `>`, `<`, `>=`, `<=`
- `IN` (requires slice/array)

### Column Selection

```go
// Fetch only specific columns
p := core.NewPaginator(
    &core.RealSession{Session: session},
    "SELECT * FROM users",  // "*" will be replaced
    core.Options{
        PageSize: 50,
        Columns: []string{"user_id", "name", "email"}, // Only fetch these
    },
)

results, token, _ := p.Next()
```

### Context-Aware Queries (Timeouts & Cancellation)

```go
import "context"

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

p := core.NewPaginator(
    &core.RealSession{Session: session},
    "SELECT * FROM large_table",
    core.Options{
        PageSize: 1000,
        Context:  ctx, // Query will be cancelled after 5 seconds
    },
)

results, token, err := p.Next()
if err != nil {
    // Handle timeout or cancellation
}
```

### Backward Navigation

```go
// Navigate forward
page1, token1, _ := p.NextWithToken("")
page2, token2, _ := p.NextWithToken(token1)
page3, token3, _ := p.NextWithToken(token2)

// Navigate backward
previousPage, prevToken, err := p.Previous(token3)
if err != nil {
    // No previous token available
}
```

**Note:** Backward navigation requires token caching. By default, the cache stores the last 10 tokens.

### Structured Logging

```go
p := core.NewPaginator(
    &core.RealSession{Session: session},
    "SELECT * FROM events",
    core.Options{
        PageSize: 100,
        Logger: func(event string, data map[string]interface{}) {
            log.Printf("[%s] %+v\n", event, data)
        },
    },
)

// Logs will be emitted for:
// - "page_fetched" (successful queries)
// - "query_failed" (errors)
// - "invalid_token" (token decoding failures)
```

### Prometheus Metrics

```go
import "github.com/AnukritiSharma1609/caspage/metrics"

// Create Prometheus collector
collector := metrics.NewPrometheusCollector()

p := core.NewPaginator(
    &core.RealSession{Session: session},
    "SELECT * FROM transactions",
    core.Options{
        PageSize: 200,
        Metrics:  collector,
    },
)

results, token, _ := p.Next()

// Metrics exposed:
// - caspage_page_fetch_duration_seconds (histogram)
// - caspage_rows_fetched_total (counter)
// - caspage_errors_total (counter by type)
// - caspage_active_tokens (gauge)
```

Expose metrics endpoint:

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

http.Handle("/metrics", promhttp.Handler())
http.ListenAndServe(":2112", nil)
```

### REST API Integration

```go
import "github.com/gin-gonic/gin"

r := gin.Default()

r.GET("/api/users", func(c *gin.Context) {
    pageToken := c.Query("pageToken")  // ?pageToken=abc123
    pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))

    p := core.NewPaginator(
        &core.RealSession{Session: session},
        "SELECT * FROM users",
        core.Options{PageSize: pageSize},
    )

    results, nextToken, err := p.NextWithToken(pageToken)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{
        "data":      results,
        "nextToken": nextToken,
        "hasMore":   nextToken != "",
    })
})

r.Run(":8080")
```

**Example API calls:**
```bash
# First page
curl http://localhost:8080/api/users?pageSize=20

# Next page
curl http://localhost:8080/api/users?pageToken=<token_from_previous_response>
```

---

## API Reference

### Core Types

#### `Paginator`

The main pagination orchestrator.

```go
type Paginator struct {
    Session  CassandraSession
    Query    string
    PageSize int
    Cache    *TokenCache
    Opts     Options
}
```

#### `Options`

Configuration options for the paginator.

```go
type Options struct {
    PageSize int                                      // Number of rows per page (default: 100)
    Filters  map[string]interface{}                   // Dynamic WHERE clauses
    Columns  []string                                 // Column selection (replaces "*")
    Context  context.Context                          // For timeouts/cancellation
    Logger   func(event string, data map[string]interface{}) // Logging hook
    Metrics  MetricsCollector                         // Metrics collection hook
}
```

### Core Methods

#### `NewPaginator`

Creates a new paginator instance.

```go
func NewPaginator(session CassandraSession, query string, opts Options) *Paginator
```

#### `Next()`

Fetches the next page (stateful). Returns results, next token, and error.

```go
func (p *Paginator) Next() ([]map[string]interface{}, string, error)
```

#### `NextWithToken(token string)`

Fetches the next page using a token (stateless). Returns results, next token, and error.

```go
func (p *Paginator) NextWithToken(token string) ([]map[string]interface{}, string, error)
```

#### `Previous(currentToken string)`

Navigates to the previous page using the token cache. Returns results, previous token, and error.

```go
func (p *Paginator) Previous(currentToken string) ([]map[string]interface{}, string, error)
```

### Token Management

#### `EncodeToken(state []byte)`

Encodes Cassandra page state into a Base64 URL-safe token.

```go
func EncodeToken(state []byte) string
```

#### `DecodeToken(token string)`

Decodes a Base64 token back into Cassandra page state.

```go
func DecodeToken(token string) ([]byte, error)
```

### Error Types

```go
var (
    ErrInvalidToken = errors.New("invalid pagination token")
    ErrQueryFailed  = errors.New("cassandra query failed")
    ErrNoPrevToken  = errors.New("no previous token available")
)
```

---

## Configuration

### Page Size

```go
core.Options{
    PageSize: 500, // Fetch 500 rows per page
}
```

**Recommendations:**
- **REST APIs:** 20-100 rows
- **Background jobs:** 500-1000 rows
- **Bulk exports:** 1000-5000 rows

### Token Cache Size

By default, the cache stores the last **10 tokens**. Modify in source if needed:

```go
// In core/paginator.go:28
Cache: NewTokenCache(20), // Keep 20 tokens instead of 10
```

### Filters

Filters are applied as `WHERE` or `AND` clauses automatically.

```go
Filters: map[string]interface{}{
    "age >=":     18,
    "status":     "active",           // Defaults to "="
    "region IN":  []string{"US", "EU"},
}
```

**Generated query:**
```sql
SELECT * FROM users WHERE age >= ? AND status = ? AND region IN (?, ?)
```

---

## Production Features

### Thread Safety

- `TokenCache` is protected by a mutex
- Safe for concurrent requests
- Each paginator instance maintains its own cache

### Error Handling

```go
results, token, err := p.Next()
if err != nil {
    switch {
    case errors.Is(err, core.ErrInvalidToken):
        // Invalid token provided
    case errors.Is(err, core.ErrQueryFailed):
        // Cassandra query failed
    case errors.Is(err, core.ErrNoPrevToken):
        // No previous page available
    }
}
```

### Observability

**Logging events:**
- `page_fetched` – Successful page retrieval
- `query_failed` – Query execution failure
- `invalid_token` – Token decoding error

**Prometheus metrics:**
- `caspage_page_fetch_duration_seconds` – Query latency
- `caspage_rows_fetched_total` – Total rows fetched
- `caspage_errors_total` – Error count by type
- `caspage_active_tokens` – Current cache size

### Performance Considerations

1. **Choose appropriate page sizes** – Larger pages = fewer round trips but higher memory usage
2. **Use column selection** – Reduce network overhead by fetching only needed columns
3. **Set query timeouts** – Use `context.WithTimeout()` to prevent hanging queries
4. **Monitor cache size** – Large caches can consume memory in high-traffic scenarios

---

## Examples

The [`examples/`](examples/) directory contains fully functional demos:

- **[`basic/main.go`](examples/basic/main.go)** – Simple pagination with Gin
- **[`restAPI/main.go`](examples/restAPI/main.go)** – Complete REST API with filters, logging, and Prometheus metrics

**Run the REST API example:**

```bash
# Start Cassandra locally (Docker)
docker run -d --name cassandra -p 9042:9042 cassandra:latest

# Run the example
cd examples/restAPI
go run main.go

# Test endpoints
curl http://localhost:8080/users?pageSize=10
curl http://localhost:8080/metrics  # Prometheus metrics
```

---

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

**Before submitting:**
- Run tests: `go test ./...`
- Run linter: `golangci-lint run`
- Add tests for new features
- Update documentation

---

## Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

Built on top of the excellent [`gocql`](https://github.com/gocql/gocql) driver.

---

## Support

- **Issues:** [GitHub Issues](https://github.com/AnukritiSharma1609/caspage/issues)
- **Discussions:** [GitHub Discussions](https://github.com/AnukritiSharma1609/caspage/discussions)
- **Documentation:** [pkg.go.dev](https://pkg.go.dev/github.com/AnukritiSharma1609/caspage)

---
