package main

import (
	"fmt"
	"os"
)

func main() {
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Ensure database directory exists
	if err := config.EnsureDatabaseDir(); err != nil {
		fmt.Printf("Error creating database directory: %v\n", err)
		os.Exit(1)
	}

	db, err := NewDB(config.DatabasePath, config.SyncUrl)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	cli := NewCLI(db)

	if len(os.Args) < 2 {
		cli.HandleCommand("ui", nil)
		return
	}

	command := os.Args[1]
	args := os.Args[2:]

	cli.HandleCommand(command, args)
}
