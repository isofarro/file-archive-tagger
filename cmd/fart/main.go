package main

import (
	"fmt"
	"os"
	"path/filepath"

	"go-fart/internal/cli"
	"go-fart/internal/database"
	"go-fart/internal/taxonomy"
)

func main() {
	var err error

	if len(os.Args) < 2 {
		fmt.Println("Usage: fart <command> [arguments]")
		os.Exit(1)
	}

	// Get the .fart database path
	dbPath := filepath.Join(".", ".fart")

	// Initialize database
	db, err := database.New(dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize managers
	taxonomyManager := taxonomy.New(db)
	cliManager := cli.New(taxonomyManager, db)

	// Handle commands
	command := os.Args[1]
	switch command {
	case "init":
		err = db.Initialize()
	case "add":
		err = cliManager.HandleAddCommand(os.Args[1:])
	case "taxonomy":
		err = cliManager.HandleTaxonomyCommand(os.Args[1:])
	case "tag":
		err = cliManager.HandleTagCommand(os.Args[1:])
	case "search":
		err = cliManager.HandleSearchCommand(os.Args[1:])
	case "check":
		err = cliManager.HandleCheckCommand(os.Args[1:])
	case "verify":
		err = cliManager.HandleVerifyCommand(os.Args[1:])
	case "normalise", "normalize":
		err = cliManager.HandleNormalizeCommand(os.Args[1:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
