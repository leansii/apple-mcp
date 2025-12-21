package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
    // Redirect standard log to stderr to avoid interfering with MCP stdio transport
    log.SetOutput(os.Stderr)

    // Create server instance
    server := mcp.NewServer(&mcp.Implementation{
        Name:    "icloud-mcp",
        Version: "1.0.0",
    }, nil)

    // Add tools
    registerTools(server)

    // Connect to transport
    transport := &mcp.StdioTransport{}

    // Connect returns a session
    log.Println("Starting iCloud MCP Server...")
    session, err := server.Connect(context.Background(), transport, nil)
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }

    // Wait for the session to close
    if err := session.Wait(); err != nil {
        log.Printf("Session closed with error: %v", err)
    }
    log.Println("Session closed.")
}
