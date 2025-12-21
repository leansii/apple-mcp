package main

import (
	"fmt"
	"log"

	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func runReadEmails(limit int) (*mcp.CallToolResult, any, error) {
    if limit <= 0 {
        limit = 10
    }

    email, err := getEnv("ICLOUD_EMAIL")
    if err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Configuration error: %v", err)}},
            IsError: true,
        }, nil, nil
    }
    password, err := getEnv("ICLOUD_PASSWORD")
    if err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Configuration error: %v", err)}},
            IsError: true,
        }, nil, nil
    }

    log.Println("Connecting to server...")

    // Connect to server
    c, err := client.DialTLS("imap.mail.me.com:993", nil)
    if err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to connect to IMAP: %v", err)}},
            IsError: true,
        }, nil, nil
    }
    defer c.Logout()

    // Login
    if err := c.Login(email, password); err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to login to IMAP: %v", err)}},
            IsError: true,
        }, nil, nil
    }

    // Select INBOX
    mbox, err := c.Select("INBOX", false)
    if err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to select INBOX: %v", err)}},
            IsError: true,
        }, nil, nil
    }

    // Get the last N messages
    from := uint32(1)
    if mbox.Messages > uint32(limit) {
        from = mbox.Messages - uint32(limit) + 1
    }
    to := mbox.Messages
    if from > to {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: "No emails found"}},
        }, nil, nil
    }

    seqset := new(imap.SeqSet)
    seqset.AddRange(from, to)

    messages := make(chan *imap.Message, 10)
    done := make(chan error, 1)
    go func() {
        done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
    }()

    var result string
    for msg := range messages {
        result += fmt.Sprintf("Subject: %s\nFrom: %v\nDate: %v\n---\n", msg.Envelope.Subject, msg.Envelope.From, msg.Envelope.Date)
    }

    if err := <-done; err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to fetch messages: %v", err)}},
            IsError: true,
        }, nil, nil
    }

    return &mcp.CallToolResult{
        Content: []mcp.Content{&mcp.TextContent{Text: result}},
    }, nil, nil
}
