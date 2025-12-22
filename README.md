# iCloud MCP Server

An [MCP](https://modelcontextprotocol.io/) server implementation for iCloud services, written in Go. This server allows AI agents to interact with your iCloud Email, and generate events for Calendar and Reminders.

## Features

| Feature | Status | Description |
| :--- | :--- | :--- |
| **Email** | ✅ Fully Supported | Send emails (SMTP) and read recent emails (IMAP). |
| **Calendar** | ⚠️ Partial | Generates valid iCalendar (`.ics`) objects for events. Direct syncing requires a specific `ICLOUD_CALDAV_URL` as auto-discovery is not fully implemented. |
| **Reminders** | ⚠️ Partial | Generates valid VTODO objects. Same limitation as Calendar for syncing. |
| **Notes** | ❌ Not Supported | iCloud Notes does not have a public API. This tool returns a placeholder message. |

## Safety & Security

### Dependencies
This project relies on well-established open-source libraries in the Go ecosystem:
*   **[emersion/go-imap](https://github.com/emersion/go-imap)**: A widely used, robust IMAP client library.
*   **[emersion/go-webdav](https://github.com/emersion/go-webdav)** & **[go-ical](https://github.com/emersion/go-ical)**: Standard libraries for handling WebDAV/CalDAV and iCalendar formats.
*   **[net/smtp](https://pkg.go.dev/net/smtp)**: The standard Go library for SMTP.

### Credentials
*   **App-Specific Passwords**: You **MUST** use an Apple App-Specific Password, not your main Apple ID password. This ensures that even if the token is compromised, your main account remains secure, and you can revoke the password at any time via [appleid.apple.com](https://appleid.apple.com).
*   **Environment Variables**: Credentials are read from environment variables (`ICLOUD_EMAIL`, `ICLOUD_PASSWORD`). Never commit these values to code or share them.

## Usage

### Prerequisites
1.  [Go](https://go.dev/dl/) installed (1.23+ recommended).
2.  An iCloud account.
3.  An **App-Specific Password** generated from [appleid.apple.com](https://appleid.apple.com).

### Installation

Clone the repository and build the server:

```bash
git clone <your-repo-url>
cd icloud-mcp
go build -o icloud-mcp
```

### Configuration

Set the following environment variables:

*   `ICLOUD_EMAIL`: Your iCloud email address (e.g., `user@icloud.com`).
*   `ICLOUD_PASSWORD`: Your App-Specific Password (format: `xxxx-xxxx-xxxx-xxxx`).
*   `ICLOUD_CALDAV_URL` (Optional): The direct URL to your specific calendar if you want to attempt direct syncing. Otherwise, the tools will generate iCal data for you to use manually.

### Running with Claude Desktop (or other MCP Clients)

Add the server to your MCP configuration (e.g., `claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "icloud": {
      "command": "/path/to/icloud-mcp",
      "env": {
        "ICLOUD_EMAIL": "your_email@icloud.com",
        "ICLOUD_PASSWORD": "your_app_specific_password"
      }
    }
  }
}
```

### Available Tools

*   `send_email`: Send an email.
    *   Args: `to`, `subject`, `body`
*   `read_emails`: Fetch recent emails.
    *   Args: `limit` (default 10)
*   `create_calendar_event`: Generate an iCalendar event.
    *   Args: `summary`, `start_time` (RFC3339), `duration_minutes`
*   `list_calendar_events`: List events (Requires `ICLOUD_CALDAV_URL`).
*   `create_reminder`: Generate a Reminder (VTODO).
    *   Args: `title`, `due_date` (optional)
*   `create_note`: (Experimental) Placeholder for Notes creation.

## License

MIT
