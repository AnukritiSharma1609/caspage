package main

import (
	"fmt"
	"log"

	"github.com/gocql/gocql"

	"github.com/AnukritiSharma1609/caspage/core"
)

// User is a strongly-typed struct matching Cassandra table columns
type User struct {
	ID    string `cql:"user_id"`
	Name  string `cql:"name"`
	Email string `cql:"email"`
}

func main() {
	// 🔹 Connect to Cassandra
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "my_keyspace"
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Failed to connect to Cassandra: %v", err)
	}
	defer session.Close()

	// 🔹 Create a paginator
	paginator := core.NewPaginator(
		&core.RealSession{Session: session},
		"SELECT * FROM users",
		core.Options{PageSize: 5},
	)

	fmt.Println("Demonstrating type-safe pagination with caspage generics")

	// 1️⃣ Fetch first page using generics
	users, nextToken, err := core.NextAs[User](paginator)
	if err != nil {
		log.Fatalf("Error fetching first page: %v", err)
	}
	printUsers("Page 1", users)

	// 2️⃣ Fetch next page using the token
	if nextToken != "" {
		moreUsers, nextToken2, err := core.NextWithTokenAs[User](paginator, nextToken)
		if err != nil {
			log.Fatalf("Error fetching next page: %v", err)
		}
		printUsers("Page 2", moreUsers)

		// 3️⃣ Fetch previous page (stateless)
		prevUsers, prevToken, err := paginator.Previous(nextToken2)
		if err != nil {
			log.Printf("Previous page error: %v", err)
		} else {
			printUsers("Previous Page (from Page 2)", convertToUsers(prevUsers))
			fmt.Println("Previous token:", prevToken)
		}
	}

	fmt.Println("Generic pagination demo completed")
}

// Helper to print user results
func printUsers(title string, users []User) {
	fmt.Printf("\n📄 %s (Total: %d)\n", title, len(users))
	for _, u := range users {
		fmt.Printf("- %s (%s)\n", u.Name, u.Email)
	}
}

// Helper to convert []map[string]interface{} → []User
// This is only for demo purposes (to show raw map results compatibility)
func convertToUsers(rows []map[string]interface{}) []User {
	users := make([]User, 0, len(rows))
	for _, row := range rows {
		u := User{}
		if id, ok := row["user_id"].(string); ok {
			u.ID = id
		}
		if name, ok := row["name"].(string); ok {
			u.Name = name
		}
		if email, ok := row["email"].(string); ok {
			u.Email = email
		}
		users = append(users, u)
	}
	return users
}
