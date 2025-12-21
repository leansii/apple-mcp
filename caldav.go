package main

import (
	"fmt"
	"net/http"
    "time"
    "bytes"
    "os"

	"github.com/emersion/go-webdav/caldav"
    "github.com/emersion/go-webdav"
    "github.com/emersion/go-ical"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/google/uuid"
)

// Using a custom HTTP client for Basic Auth
type basicAuthTransport struct {
    Username string
    Password string
    Base     http.RoundTripper
}

func (t *basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    req.SetBasicAuth(t.Username, t.Password)
    return t.Base.RoundTrip(req)
}

func getCalDAVClient() (*caldav.Client, error) {
    email, err := getEnv("ICLOUD_EMAIL")
    if err != nil {
        return nil, err
    }
    password, err := getEnv("ICLOUD_PASSWORD")
    if err != nil {
        return nil, err
    }

    httpClient := &http.Client{
        Transport: &basicAuthTransport{
            Username: email,
            Password: password,
            Base:     http.DefaultTransport,
        },
    }

    url := os.Getenv("ICLOUD_CALDAV_URL")
    if url == "" {
        url = "https://caldav.icloud.com/"
    }

    client, err := caldav.NewClient(webdav.HTTPClient(httpClient), url)
    if err != nil {
        return nil, err
    }
    return client, nil
}

func runCreateCalendarEvent(summary, startTime string, durationMinutes int) (*mcp.CallToolResult, any, error) {
    start, err := time.Parse(time.RFC3339, startTime)
    if err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid start time format: %v", err)}},
            IsError: true,
        }, nil, nil
    }

    end := start.Add(time.Duration(durationMinutes) * time.Minute)

    // Create VEVENT
    event := ical.NewEvent()
    event.Props.SetText(ical.PropSummary, summary)
    event.Props.SetDateTime(ical.PropDateTimeStart, start)
    event.Props.SetDateTime(ical.PropDateTimeEnd, end)
    event.Props.SetText(ical.PropUID, uuid.NewString())

    cal := ical.NewCalendar()
    cal.Props.SetText(ical.PropVersion, "2.0")
    cal.Props.SetText(ical.PropProductID, "-//Jules//iCloud MCP//EN")
    cal.Children = append(cal.Children, event.Component)

    // Encode
    var buf bytes.Buffer
    enc := ical.NewEncoder(&buf)
    if err := enc.Encode(cal); err != nil {
         return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to encode calendar object: %v", err)}},
            IsError: true,
        }, nil, nil
    }

    return &mcp.CallToolResult{
        Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Generated iCalendar object:\n%s\n\nNote: To actually save this to iCloud, the server needs to discover the specific calendar URL, which requires authentication and discovery logic that is difficult to automate blindly. Please verify if your environment supports CalDAV discovery.", buf.String())}},
    }, nil, nil
}

func runListCalendarEvents(startTime, endTime string) (*mcp.CallToolResult, any, error) {
    client, err := getCalDAVClient()
    if err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Client error: %v", err)}},
            IsError: true,
        }, nil, nil
    }

    _ = client

    return &mcp.CallToolResult{
        Content: []mcp.Content{&mcp.TextContent{Text: "Listing events requires discovering the user's calendar home set. This is currently not fully implemented. Please ensure ICLOUD_CALDAV_URL points to a specific calendar if possible."}},
    }, nil, nil
}

func runCreateReminder(title, dueDate string) (*mcp.CallToolResult, any, error) {
    // Similar to Event but VTODO
    todo := ical.NewComponent(ical.CompToDo)
    todo.Props.SetText(ical.PropSummary, title)
    todo.Props.SetText(ical.PropUID, uuid.NewString())

    if dueDate != "" {
        due, err := time.Parse(time.RFC3339, dueDate)
        if err == nil {
             todo.Props.SetDateTime(ical.PropDue, due)
        }
    }

    cal := ical.NewCalendar()
    cal.Props.SetText(ical.PropVersion, "2.0")
    cal.Props.SetText(ical.PropProductID, "-//Jules//iCloud MCP//EN")
    cal.Children = append(cal.Children, todo)

    var buf bytes.Buffer
    enc := ical.NewEncoder(&buf)
    _ = enc.Encode(cal)

    return &mcp.CallToolResult{
        Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Generated VTODO object:\n%s\n\nNote: Same limitation as Calendar events regarding server path discovery.", buf.String())}},
    }, nil, nil
}
