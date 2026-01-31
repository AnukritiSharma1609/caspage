[![Go Build](https://github.com/AnukritiSharma1609/caspage/actions/workflows/ci.yml/badge.svg)](https://github.com/AnukritiSharma1609/caspage/actions)
[![codecov](https://codecov.io/gh/AnukritiSharma1609/caspage/branch/main/graph/badge.svg)](https://codecov.io/gh/AnukritiSharma1609/caspage)
[![Go Reference](https://pkg.go.dev/badge/github.com/AnukritiSharma1609/caspage.svg)](https://pkg.go.dev/github.com/AnukritiSharma1609/caspage)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)


# caspage

`caspage` is a lightweight, high-performance pagination library for Cassandra in Go, built on top of `gocql`.

It provides stateful and stateless pagination, query filters, context handling, metrics hooks, and logging — all without changing your underlying Cassandra schema or application logic.

Think of it as gocql pagination done right: simpler, safer, and more production-ready.

---

## Why caspage?

Pagination in Cassandra using `gocql` has always been tricky:

| Problem with raw gocql                              | How caspage solves it                                             |
|-----------------------------------------------------|--------------------------------------------------------------------|
| No direct API for cursor-based pagination           | Provides both stateful (`Next`) and stateless (`NextWithToken`) pagination |
| Page state tokens aren’t REST-safe                  | Encodes and decodes them into portable Base64 tokens              |
| Requires manual handling of iterators               | Automatically manages page tokens and iterator lifecycle          |
| No previous page or cache support                   | Built-in token cache with `Previous()` navigation                 |
| No built-in metrics, logging, or filters            | Ships with Prometheus hooks, structured logging, and query filters|
| No context awareness                                | Supports `context.Context` for cancellation and timeouts          |

---

## Features

- Simple API – paginate results using just `Next()` or `NextWithToken()`
- Stateless pagination – tokens can be safely shared via REST APIs
- Bidirectional navigation – move forward and backward between pages
- Filter support – add dynamic `WHERE` clauses with operators (`>`, `<`, `IN`, etc.)
- Token cache – keep track of visited tokens in memory
- Context-aware queries – use `context.Context` for safe cancellations
- Metrics hooks – plug in Prometheus (or any custom collector) easily
- Structured logging – log query performance and pagination details
- Drop-in compatible with `gocql` – no schema changes or driver hacks needed

---

## Installation

```bash
go get github.com/AnukritiSharma1609/caspage








