package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/AnukritiSharma1609/caspage/core"
	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
)

func main() {
	r := gin.Default()

	// Cassandra connection setup
	cluster := gocql.NewCluster("127.0.0.1") // or your host
	cluster.Keyspace = "your_keyspace"

	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Failed to connect to Cassandra: %v", err)
	}
	defer session.Close()

	// Create paginator (no cache needed now)
	paginator := core.NewPaginator(&core.RealSession{Session: session},
		"SELECT * FROM users",
		core.Options{PageSize: 10})

	// ------------------------------
	// Example 1 — Next() (start fresh)
	// ------------------------------
	r.GET("/users/basic", func(c *gin.Context) {
		results, nextToken, err := paginator.Next()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data":       results,
			"next_token": nextToken,
		})
	})

	// ------------------------------
	// Example 2 — NextWithToken() (stateless forward)
	// ------------------------------
	r.GET("/users", func(c *gin.Context) {
		token := c.Query("pageToken")
		results, nextToken, err := paginator.NextWithToken(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data":       results,
			"next_token": nextToken,
		})
	})

	// ------------------------------
	// Example 3 — Previous() (stateless backward)
	// ------------------------------
	r.GET("/users/previous", func(c *gin.Context) {
		token := c.Query("pageToken")
		results, prevToken, err := paginator.Previous(token)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data":         results,
			"previousPage": prevToken,
		})
	})

	// ------------------------------
	// Example 4 — Demo (simulate navigation)
	// ------------------------------
	r.GET("/demo", func(c *gin.Context) {
		results1, next1, _ := paginator.NextWithToken("")
		results2, next2, _ := paginator.NextWithToken(next1)
		prevResults, prevToken, _ := paginator.Previous(next2)

		fmt.Printf("Page 1 next token: %s\n", next1)
		fmt.Printf("Page 2 next token: %s\n", next2)
		fmt.Printf("Went back to previous page: %s\n", prevToken)
		fmt.Printf("Page 1 rows: %d, Page 2 rows: %d, Previous rows: %d\n",
			len(results1), len(results2), len(prevResults))

		c.JSON(http.StatusOK, gin.H{
			"page1_count": len(results1),
			"page2_count": len(results2),
			"prev_count":  len(prevResults),
		})
	})

	fmt.Println("Server running at http://localhost:8080")
	fmt.Println("➡ http://localhost:8080/users/basic")
	fmt.Println("➡ http://localhost:8080/users?pageToken=<token>")
	fmt.Println("➡ http://localhost:8080/users/previous?pageToken=<token>")
	fmt.Println("➡ http://localhost:8080/demo")

	r.Run(":8080")
}
