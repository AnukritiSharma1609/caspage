package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/AnukritiSharma1609/caspage/core"
	"github.com/AnukritiSharma1609/caspage/metrics"
)

func main() {
	// Initialize Gin
	r := gin.Default()

	// Connect to Cassandra
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "merchant_platform"
	cluster.Consistency = gocql.Quorum
	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Initialize Prometheus collector
	collector := metrics.NewPrometheusCollector()

	// ------------------------------
	// /users endpoint with pagination, filters, logging, and metrics
	// ------------------------------
	r.GET("/users", func(c *gin.Context) {
		pageToken := c.Query("pageToken")
		pageSizeStr := c.DefaultQuery("pageSize", "20")
		pageSize, _ := strconv.Atoi(pageSizeStr)

		// Example filter parsing from query params:
		// ?filters=age>25,regionIN(US|CA),active=true
		filterStr := c.Query("filters")
		filters := parseFilters(filterStr)

		// Use context to apply 5s timeout per request
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		p := core.NewPaginator(&core.RealSession{Session: session}, "SELECT * FROM user_role_mapping_v2", core.Options{
			PageSize: pageSize,
			Context:  ctx,
			Filters:  filters,
			Columns:  []string{"user_id", "app_data", "role_ids", "name", "count"},
			Metrics:  collector,
			Logger: func(event string, data map[string]interface{}) {
				log.Printf("[LOG] %s: %+v\n", event, data)
			},
		})

		results, nextToken, err := p.NextWithToken(pageToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":       results,
			"next_token": nextToken,
			"filters":    filters,
		})
	})

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Start server
	r.Run(":8080")
}

// --------------------------------------------
// Helper: Parse query param filters dynamically
// --------------------------------------------
// Example: "age>25,regionIN(US|CA),active=true"
// Converts to: map[string]interface{}{"age >": 25, "region IN": []string{"US", "CA"}, "active": true}
func parseFilters(param string) map[string]interface{} {
	if param == "" {
		return nil
	}

	filters := make(map[string]interface{})
	pairs := strings.Split(param, ",")

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		switch {
		case strings.Contains(pair, "IN("):
			// e.g., regionIN(US|CA)
			parts := strings.SplitN(pair, "IN(", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0]) + " IN"
				values := strings.TrimSuffix(parts[1], ")")
				filters[key] = strings.Split(values, "|")
			}

		case strings.Contains(pair, ">="):
			kv := strings.SplitN(pair, ">=", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0]) + " >="
				filters[key] = parseValue(kv[1])
			}

		case strings.Contains(pair, "<="):
			kv := strings.SplitN(pair, "<=", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0]) + " <="
				filters[key] = parseValue(kv[1])
			}

		case strings.Contains(pair, ">"):
			kv := strings.SplitN(pair, ">", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0]) + " >"
				filters[key] = parseValue(kv[1])
			}

		case strings.Contains(pair, "<"):
			kv := strings.SplitN(pair, "<", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0]) + " <"
				filters[key] = parseValue(kv[1])
			}

		case strings.Contains(pair, "="):
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				filters[key] = parseValue(kv[1])
			}
		}
	}
	return filters
}

// parseValue converts string to int, bool, or keeps as string
func parseValue(val string) interface{} {
	val = strings.TrimSpace(val)
	if val == "" {
		return nil
	}
	// try int
	if i, err := strconv.Atoi(val); err == nil {
		return i
	}
	// try bool
	if val == "true" {
		return true
	}
	if val == "false" {
		return false
	}
	// fallback string
	return val
}
