package main

import (
	"fmt"
	"log"

	"github.com/AnukritiSharma1609/caspage/core"
	"github.com/gocql/gocql"
)

func main() {
	cluster := gocql.NewCluster("127.0.0.1") // or your Cassandra host
	cluster.Keyspace = "your_keyspace"
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("unable to connect to Cassandra: %v", err)
	}
	defer session.Close()

	opts := core.Options{PageSize: 50}
	p := core.NewPaginator(session, "SELECT * FROM users", opts)

	results, next, err := p.Next()
	if err != nil {
		log.Fatalf("error fetching page: %v", err)
	}

	fmt.Printf("Fetched %d rows. Next token: %v\n", len(results), next != "")
}
