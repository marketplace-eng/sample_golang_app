package main

import (
	"context"
	"fmt"
	"os"
	"sample_app/internal/database"
	"sample_app/internal/server"
)

func main() {
	ctx := context.Background()
	db, err := database.OpenDB()
	if err != nil {
		fmt.Printf("Unable to connect to database. Exiting.")
		os.Exit(1)
	}
	defer db.Close()

	// Ping to ensure DB connection works
	var greeting string
	err = db.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Startup QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)

	server.StartServer(ctx, db)
}
