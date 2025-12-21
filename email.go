package main

import (
	"fmt"
	"net/smtp"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

func runSendEmail(to, subject, body string) (*mcp.CallToolResult, any, error) {
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

    smtpHost := "smtp.mail.me.com"
    smtpPort := "587"

    // Message
    msg := []byte("To: " + to + "\r\n" +
        "Subject: " + subject + "\r\n" +
        "\r\n" +
        body + "\r\n")

    // Authentication
    auth := smtp.PlainAuth("", email, password, smtpHost)

    // Sending email
    err = smtp.SendMail(smtpHost+":"+smtpPort, auth, email, []string{to}, msg)
    if err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to send email: %v", err)}},
            IsError: true,
        }, nil, nil
    }

    return &mcp.CallToolResult{
        Content: []mcp.Content{&mcp.TextContent{Text: "Email sent successfully"}},
    }, nil, nil
}
