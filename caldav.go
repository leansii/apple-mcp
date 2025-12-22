package main

import (
	"fmt"
	"net/http"
    "time"
    "bytes"
    "os"
    "context"

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

func getCalDAVClient(urlEnv string) (*caldav.Client, error) {
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

    url := os.Getenv(urlEnv)
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
    uid := uuid.NewString()

    // Create VEVENT
    event := ical.NewEvent()
    event.Props.SetText(ical.PropSummary, summary)
    event.Props.SetDateTime(ical.PropDateTimeStart, start)
    event.Props.SetDateTime(ical.PropDateTimeEnd, end)
    event.Props.SetText(ical.PropUID, uid)

    cal := ical.NewCalendar()
    cal.Props.SetText(ical.PropVersion, "2.0")
    cal.Props.SetText(ical.PropProductID, "-//Jules//iCloud MCP//EN")
    cal.Children = append(cal.Children, event.Component)

    var buf bytes.Buffer
    enc := ical.NewEncoder(&buf)
    if err := enc.Encode(cal); err != nil {
         return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to encode calendar object: %v", err)}},
            IsError: true,
        }, nil, nil
    }

    // Tentative: If we wanted to upload, we would need to construct the path.
    // client.PutCalendarObject(context.Background(), path, &buf, options)
    // But without knowing the correct collection path, this often fails.

    return &mcp.CallToolResult{
        Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Generated iCalendar object:\n%s\n\nNote: Event generated locally. To upload, manual action or precise CalDAV URL configuration is required.", buf.String())}},
    }, nil, nil
}

func runListCalendarEvents(startTime, endTime string) (*mcp.CallToolResult, any, error) {
    // Check if ICLOUD_CALDAV_URL is set before trying to connect
    if os.Getenv("ICLOUD_CALDAV_URL") == "" {
         return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: "Please configure ICLOUD_CALDAV_URL to a specific calendar to list events."}},
        }, nil, nil
    }

    client, err := getCalDAVClient("ICLOUD_CALDAV_URL")
    if err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Client error: %v", err)}},
            IsError: true,
        }, nil, nil
    }

    // Construct query for VEVENT
    query := &caldav.CalendarQuery{
        CompRequest: caldav.CalendarCompRequest{
            Name: "VCALENDAR",
            Comps: []caldav.CalendarCompRequest{
                {
                    Name: "VEVENT",
                    Props: []string{"SUMMARY", "DTSTART", "DTEND", "UID", "DESCRIPTION", "LOCATION"},
                },
            },
        },
        CompFilter: caldav.CompFilter{
            Name: "VCALENDAR",
            Comps: []caldav.CompFilter{
                {
                    Name: "VEVENT",
                },
            },
        },
    }

    // Add time range filter if possible
    start, errS := time.Parse(time.RFC3339, startTime)
    end, errE := time.Parse(time.RFC3339, endTime)

    if errS == nil && errE == nil {
         // Modify the VEVENT filter
         query.CompFilter.Comps[0].Start = start
         query.CompFilter.Comps[0].End = end
    }

    objs, err := client.QueryCalendar(context.Background(), "", query)
    if err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to query events: %v. Ensure ICLOUD_CALDAV_URL points to a Calendar collection.", err)}},
            IsError: true,
        }, nil, nil
    }

    var result string
    for _, obj := range objs {
        result += fmt.Sprintf("Event found at: %s\n", obj.Path)

        if obj.Data != nil {
             for _, child := range obj.Data.Children {
                 if child.Name == "VEVENT" {
                     summary := child.Props.Get(ical.PropSummary)
                     if summary != nil {
                         result += fmt.Sprintf("  Summary: %s\n", summary.Value)
                     }
                     dtstart := child.Props.Get(ical.PropDateTimeStart)
                     if dtstart != nil {
                         result += fmt.Sprintf("  Start: %s\n", dtstart.Value)
                     }
                 }
             }
        }
    }

    if result == "" {
        result = "No events found in the specified range."
    }

    return &mcp.CallToolResult{
        Content: []mcp.Content{&mcp.TextContent{Text: result}},
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
        Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Generated VTODO object:\n%s\n", buf.String())}},
    }, nil, nil
}

func runListReminders() (*mcp.CallToolResult, any, error) {
    // Check if separate Reminders URL is set, otherwise try default
    if os.Getenv("ICLOUD_REMINDERS_URL") == "" && os.Getenv("ICLOUD_CALDAV_URL") == "" {
         return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: "Please configure ICLOUD_REMINDERS_URL (or ICLOUD_CALDAV_URL) to list reminders."}},
        }, nil, nil
    }

    client, err := getCalDAVClient("ICLOUD_REMINDERS_URL")
    if err != nil {
        return &mcp.CallToolResult{
             Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Client error: %v", err)}},
             IsError: true,
        }, nil, nil
    }

    // We would Query for VTODO
    query := &caldav.CalendarQuery{
        CompRequest: caldav.CalendarCompRequest{
            Name: "VCALENDAR",
            Comps: []caldav.CalendarCompRequest{
                {
                    Name: "VTODO",
                    Props: []string{"SUMMARY", "DUE", "STATUS", "UID"},
                },
            },
        },
        CompFilter: caldav.CompFilter{
            Name: "VCALENDAR",
            Comps: []caldav.CompFilter{
                {
                    Name: "VTODO",
                },
            },
        },
    }

    // Execute query
    objs, err := client.QueryCalendar(context.Background(), "", query)
    if err != nil {
         return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to query reminders: %v. Ensure the URL points to a Reminders collection.", err)}},
            IsError: true,
        }, nil, nil
    }

    var result string
    for _, obj := range objs {
        result += fmt.Sprintf("Found object at %s\n", obj.Path)
        if obj.Data != nil {
             for _, child := range obj.Data.Children {
                 if child.Name == "VTODO" {
                     summary := child.Props.Get(ical.PropSummary)
                     if summary != nil {
                         result += fmt.Sprintf("  Summary: %s\n", summary.Value)
                     }
                     status := child.Props.Get("STATUS")
                     if status != nil {
                         result += fmt.Sprintf("  Status: %s\n", status.Value)
                     }
                 }
             }
        }
    }

    if result == "" {
        result = "No reminders found."
    }

    return &mcp.CallToolResult{
        Content: []mcp.Content{&mcp.TextContent{Text: result}},
    }, nil, nil
}
