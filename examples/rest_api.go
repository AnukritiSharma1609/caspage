package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/AnukritiSharma1609/caspage/core"
)

func main() {
	r := gin.Default()

	// Cassandra connection setup
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "your_keyspace"
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Failed to connect to Cassandra: %v", err)
	}
	defer session.Close()

	// ------------------------------
	// 🥇 Example 1 — Next() (Stateful)
	// ------------------------------
	r.GET("/users/basic", func(c *gin.Context) {
		p := core.NewPaginator(session, "SELECT * FROM users", core.Options{PageSize: 50})

		results, nextToken, err := p.Next()
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
	// 🥈 Example 2 — NextWithToken() (Stateless forward)
	// ------------------------------
	r.GET("/users", func(c *gin.Context) {
		token := c.Query("pageToken")

		p := core.NewPaginator(session, "SELECT * FROM users", core.Options{PageSize: 50})
		results, nextToken, err := p.NextWithToken(token)
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
	// 🥉 Example 3 — Previous() (Backward navigation)
	// ------------------------------
	r.GET("/users/previous", func(c *gin.Context) {
		token := c.Query("pageToken")

		p := core.NewPaginator(session, "SELECT * FROM users", core.Options{PageSize: 50})
		results, prevToken, err := p.Previous(token)
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
	// 🧪 Example 4 — CLI simulation (optional testing)
	// ------------------------------
	r.GET("/demo", func(c *gin.Context) {
		p := core.NewPaginator(session, "SELECT * FROM users", core.Options{PageSize: 50})

		// Forward pages
		results, next1, _ := p.NextWithToken("")
		results2, next2, _ := p.NextWithToken(next1)

		// Backward one page
		prevResults, prevToken, _ := p.Previous(next2)

		fmt.Printf("Page 1 next token: %s\n", next1)
		fmt.Printf("Page 2 next token: %s\n", next2)
		fmt.Printf("Went back to previous page: %s\n", prevToken)
		fmt.Printf("Page 1 rows: %d, Page 2 rows: %d, Previous rows: %d\n",
			len(results), len(results2), len(prevResults))

		c.JSON(http.StatusOK, gin.H{
			"page1_count": len(results),
			"page2_count": len(results2),
			"prev_count":  len(prevResults),
		})
	})

	fmt.Println("🚀 Server running at http://localhost:8080")
	fmt.Println("Try: ")
	fmt.Println("➡ http://localhost:8080/users/basic        (Next)")
	fmt.Println("➡ http://localhost:8080/users              (NextWithToken)")
	fmt.Println("➡ http://localhost:8080/users/previous     (Previous)")
	fmt.Println("➡ http://localhost:8080/demo               (Local test simulation)")

	r.Run(":8080")
}
