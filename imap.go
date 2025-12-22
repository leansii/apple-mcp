package main

import (
	"fmt"
	"log"

	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func runReadEmails(limit int) (*mcp.CallToolResult, any, error) {
    return fetchMessages("INBOX", limit)
}

func runReadNotes(limit int) (*mcp.CallToolResult, any, error) {
    // Attempt to read from "Notes" mailbox
    result, _, err := fetchMessages("Notes", limit)
    if err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to read notes: %v. \n\nNote: Modern iCloud Notes are not accessible via IMAP. This tool only retrieves legacy notes.", err)}},
            IsError: true,
        }, nil, nil
    }

    return &mcp.CallToolResult{
        Content: []mcp.Content{&mcp.TextContent{Text: "Legacy Notes (Modern iCloud Notes are not accessible via IMAP):\n\n" + result.Content[0].(*mcp.TextContent).Text}},
    }, nil, nil
}

func fetchMessages(mailbox string, limit int) (*mcp.CallToolResult, any, error) {
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

    log.Printf("Connecting to IMAP server to fetch %s...", mailbox)

    c, err := client.DialTLS("imap.mail.me.com:993", nil)
    if err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to connect to IMAP: %v", err)}},
            IsError: true,
        }, nil, nil
    }
    defer c.Logout()

    if err := c.Login(email, password); err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to login to IMAP: %v", err)}},
            IsError: true,
        }, nil, nil
    }

    mbox, err := c.Select(mailbox, false)
    if err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to select mailbox '%s': %v. It might not exist.", mailbox, err)}},
            IsError: true,
        }, nil, nil
    }

    from := uint32(1)
    if mbox.Messages > uint32(limit) {
        from = mbox.Messages - uint32(limit) + 1
    }
    to := mbox.Messages
    if from > to {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: "No messages found"}},
        }, nil, nil
    }

    seqset := new(imap.SeqSet)
    seqset.AddRange(from, to)

    messages := make(chan *imap.Message, 10)
    done := make(chan error, 1)
    go func() {
        done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, "BODY[TEXT]"}, messages)
    }()

    var result string
    for msg := range messages {
        result += fmt.Sprintf("Subject: %s\nDate: %v\n---\n", msg.Envelope.Subject, msg.Envelope.Date)
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
