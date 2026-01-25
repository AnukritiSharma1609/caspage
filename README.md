[![codecov](https://codecov.io/gh/AnukritiSharma1609/caspage/branch/main/graph/badge.svg?token=YOUR_CODECOV_TOKEN)](https://codecov.io/gh/AnukritiSharma1609/caspage)

# caspage

> A developer-friendly Go library for efficient and stateless pagination in Cassandra, built on top of gocql.

## 🚀 Overview
`caspage` simplifies Cassandra pagination by providing:
- Clean APIs: `Next()` and `Previous()`
- Stateless page tokens for REST/gRPC services
- Optional backward navigation using cached tokens
- Prometheus metrics for observability

## 📦 Installation

### Examples
- 🧩 **Basic Pagination:** [examples/basic](examples/basic/main.go)
- ⚙️ **REST API with Filters, Metrics & Logging:** [examples/rest_api](examples/rest_api/main.go)

