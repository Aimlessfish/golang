# Steam Bot - Discord Integration

A Discord bot for managing multiple Steam browser sessions with automation capabilities.

## Project Structure

```
csrep/
├── bot/                    # Core bot library
│   ├── session.go         # Session management with Playwright
│   ├── manager.go         # Session manager with timeout support
│   ├── features.go        # Feature interface
│   └── features_impl.go   # Built-in features (navigate, screenshot, etc.)
├── cmd/                    # Command implementations
│   ├── main.go            # CLI and Discord entry points
│   └── discord.go         # Discord bot implementation
├── main.go                # Legacy standalone (deprecated)
├── manager.go             # Legacy manager (deprecated)
└── session.go             # Legacy session (deprecated)
```

## Features

### Session Management
- **Auto-generated session IDs** - No manual ID assignment needed
- **Browser automation** - Chromium-based with Playwright
- **Blocking login flow** - Sessions only become active after successful login
- **Timeout support** - Auto-remove sessions after specified duration
- **Extend timeout** - Keep sessions alive longer

### Bot Features (Modular)
- **Navigate** - Send sessions to specific URLs
- **Screenshot** - Capture current page state
- **Open Profile** - Quick Steam profile access
- **Status** - Get session information

### Interfaces
- **CLI Mode** - Interactive terminal interface
- **Discord Mode** - Discord bot commands (framework ready)
- **HTTP API** - Each session exposes REST endpoints

## Installation

```bash
# Install dependencies
go mod download

# Install Playwright browsers
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install chromium

# Build
go build -o csrep cmd/main.go cmd/discord.go
```

## Usage

### CLI Mode

```bash
./csrep server
```

**Commands:**
```bash
# Add session (no timeout)
> add myusername

# Add session with 30 minute auto-remove
> add myusername 30

# List all sessions
> list

# Extend session timeout by 15 minutes
> extend session-1 15

# Remove session
> remove session-1

# Exit
> quit
```

### Discord Mode (Framework)

```bash
./csrep discord
```

**Planned Commands:**
- `!steam add <username> [timeout]` - Create new session
- `!steam remove <session-id>` - Remove session
- `!steam list` - List all sessions
- `!steam screenshot <session-id>` - Get screenshot as attachment
- `!steam navigate <session-id> <url>` - Navigate to URL

## How It Works

### Session Lifecycle

1. **Creation** - User runs `add username`
2. **Browser Launch** - Chromium opens to Steam login
3. **Login Monitoring** - Detects successful login automatically
4. **Activation** - HTTP API server starts on unique port
5. **Ready** - Session accepts automation commands
6. **Timeout** (optional) - Auto-removes after duration
7. **Removal** - Browser closes, resources cleaned up

### Auto-Login Detection

The bot monitors for:
- URL changes away from login pages
- Presence of Steam account UI elements
- Successful navigation to Steam community/store

No manual "login complete" call needed!

### Feature System

Features are modular and implement:
```go
type Feature interface {
    Name() string
    Execute(session *Session, args map[string]interface{}) (interface{}, error)
}
```

Add custom features by implementing this interface.

## API Endpoints (Per Session)

Each session runs on its own port (starting from 8080):

```bash
# Get session status
GET http://localhost:8080/status

# Get screenshot
GET http://localhost:8080/action/get-screenshot

# Navigate to URL
POST http://localhost:8080/action/navigate
Content-Type: application/json
{"url": "https://steamcommunity.com/market"}

# Open profile
POST http://localhost:8080/action/open-profile
```

## Next Steps for Discord Integration

To complete Discord bot:

1. Add `github.com/bwmarrin/discordgo` dependency
2. Implement command parser in `cmd/discord.go`
3. Handle Discord message events
4. Upload screenshots as Discord attachments
5. Add permission checking
6. Rate limiting per Discord user

## Requirements

- Go 1.22+
- Playwright for Go
- Chromium (via Playwright)
- Linux with X11/Wayland for GUI browser
