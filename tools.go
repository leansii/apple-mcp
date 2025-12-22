package main

import (
	"context"
    "os"
    "fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerTools(server *mcp.Server) {
    // Email Tools
    mcp.AddTool(server, &mcp.Tool{
        Name: "send_email",
        Description: "Send an email using iCloud SMTP. Requires ICLOUD_EMAIL and ICLOUD_PASSWORD (app-specific) environment variables.",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "to": map[string]any{"type": "string", "description": "Recipient email address"},
                "subject": map[string]any{"type": "string", "description": "Email subject"},
                "body": map[string]any{"type": "string", "description": "Email body content"},
            },
            "required": []string{"to", "subject", "body"},
        },
    }, handleSendEmail)

    mcp.AddTool(server, &mcp.Tool{
        Name: "read_emails",
        Description: "Read recent emails from iCloud IMAP. Requires ICLOUD_EMAIL and ICLOUD_PASSWORD (app-specific) environment variables.",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "limit": map[string]any{"type": "integer", "description": "Number of emails to fetch (default 10)"},
            },
        },
    }, handleReadEmails)

    // Calendar Tools
    mcp.AddTool(server, &mcp.Tool{
        Name: "create_calendar_event",
        Description: "Create a calendar event. Requires ICLOUD_EMAIL and ICLOUD_PASSWORD (app-specific).",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "summary": map[string]any{"type": "string", "description": "Event title/summary"},
                "start_time": map[string]any{"type": "string", "description": "Start time in RFC3339 format (e.g. 2023-10-27T10:00:00Z)"},
                "duration_minutes": map[string]any{"type": "integer", "description": "Duration in minutes"},
            },
            "required": []string{"summary", "start_time", "duration_minutes"},
        },
    }, handleCreateCalendarEvent)

    mcp.AddTool(server, &mcp.Tool{
        Name: "list_calendar_events",
        Description: "List calendar events. Requires ICLOUD_CALDAV_URL pointing to a specific calendar collection.",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "start_time": map[string]any{"type": "string", "description": "Start time range (RFC3339)"},
                "end_time": map[string]any{"type": "string", "description": "End time range (RFC3339)"},
            },
            "required": []string{"start_time", "end_time"},
        },
    }, handleListCalendarEvents)

    // Reminder Tools
    mcp.AddTool(server, &mcp.Tool{
        Name: "create_reminder",
        Description: "Create a reminder. Requires ICLOUD_EMAIL and ICLOUD_PASSWORD (app-specific).",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "title": map[string]any{"type": "string", "description": "Reminder title"},
                "due_date": map[string]any{"type": "string", "description": "Due date in RFC3339 format (optional)"},
            },
            "required": []string{"title"},
        },
    }, handleCreateReminder)

    mcp.AddTool(server, &mcp.Tool{
        Name: "list_reminders",
        Description: "List reminders. Requires ICLOUD_REMINDERS_URL (or ICLOUD_CALDAV_URL) pointing to a reminders collection.",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{},
        },
    }, handleListReminders)

    // Notes Tools
    mcp.AddTool(server, &mcp.Tool{
        Name: "read_notes",
        Description: "Read Notes from the 'Notes' IMAP folder. Only works for legacy notes.",
        InputSchema: map[string]any{
             "type": "object",
             "properties": map[string]any{
                 "limit": map[string]any{"type": "integer", "description": "Number of notes to fetch (default 10)"},
             },
        },
    }, handleReadNotes)

    mcp.AddTool(server, &mcp.Tool{
        Name: "create_note",
        Description: "Create a note (Experimental/Not fully supported).",
        InputSchema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "content": map[string]any{"type": "string", "description": "Note content"},
            },
            "required": []string{"content"},
        },
    }, handleCreateNote)
}

func getEnv(key string) (string, error) {
    val := os.Getenv(key)
    if val == "" {
        return "", fmt.Errorf("environment variable %s is not set", key)
    }
    return val, nil
}

func handleCreateNote(ctx context.Context, req *mcp.CallToolRequest, args struct {
    Content string `json:"content"`
}) (*mcp.CallToolResult, any, error) {
    return &mcp.CallToolResult{
        Content: []mcp.Content{
            &mcp.TextContent{Text: "Note creation is not fully implemented yet. Use the iCloud web interface."},
        },
        IsError: true,
    }, nil, nil
}

func handleReadNotes(ctx context.Context, req *mcp.CallToolRequest, args struct {
    Limit int `json:"limit"`
}) (*mcp.CallToolResult, any, error) {
    return runReadNotes(args.Limit)
}

// Placeholders for other handlers to allow compilation
func handleSendEmail(ctx context.Context, req *mcp.CallToolRequest, args struct {
    To string `json:"to"`
    Subject string `json:"subject"`
    Body string `json:"body"`
}) (*mcp.CallToolResult, any, error) {
    return runSendEmail(args.To, args.Subject, args.Body)
}

func handleReadEmails(ctx context.Context, req *mcp.CallToolRequest, args struct {
    Limit int `json:"limit"`
}) (*mcp.CallToolResult, any, error) {
    return runReadEmails(args.Limit)
}

func handleCreateCalendarEvent(ctx context.Context, req *mcp.CallToolRequest, args struct {
    Summary string `json:"summary"`
    StartTime string `json:"start_time"`
    DurationMinutes int `json:"duration_minutes"`
}) (*mcp.CallToolResult, any, error) {
    return runCreateCalendarEvent(args.Summary, args.StartTime, args.DurationMinutes)
}

func handleListCalendarEvents(ctx context.Context, req *mcp.CallToolRequest, args struct {
    StartTime string `json:"start_time"`
    EndTime string `json:"end_time"`
}) (*mcp.CallToolResult, any, error) {
    return runListCalendarEvents(args.StartTime, args.EndTime)
}

func handleCreateReminder(ctx context.Context, req *mcp.CallToolRequest, args struct {
    Title string `json:"title"`
    DueDate string `json:"due_date"`
}) (*mcp.CallToolResult, any, error) {
    return runCreateReminder(args.Title, args.DueDate)
}

func handleListReminders(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
    return runListReminders()
}
