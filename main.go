package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/server"
	"github.com/venkytv/inventory-mgmt/db"
	"github.com/venkytv/inventory-mgmt/tools"
)

func main() {
	log.SetOutput(os.Stderr)

	dbPath := os.Getenv("INVENTORY_DB_PATH")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("cannot determine home directory: %v", err)
		}
		dbPath = filepath.Join(home, ".inventory", "inventory.db")
	}

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("opening database: %v", err)
	}
	defer database.Close()

	s := server.NewMCPServer(
		"inventory",
		"0.2.0",
		server.WithToolCapabilities(true),
	)
	tools.RegisterTools(s, database)

	log.Printf("inventory MCP server starting (db: %s)", dbPath)
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
