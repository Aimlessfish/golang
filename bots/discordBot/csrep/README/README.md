# Steam Bot Session Manager (csrep)

A multi-session bot manager that spawns multiple Steam sessions with browser automation. Each session runs in its own Chromium browser instance, listening on individual ports for action triggers.

## Features

- **Auto-generated session IDs** - No need to manually specify session IDs
- **Browser automation** - Uses Playwright to control real Chrome browsers
- **Semi-automated login** - Browser opens at Steam login, you complete the login manually
- **Multiple concurrent sessions** - Each session has its own browser and HTTP server
- **REST API endpoints** - Trigger actions, navigate, get screenshots
- **Interactive CLI** - Easy session management

## Architecture

- **Session Manager**: Orchestrates multiple bot sessions
- **Individual Sessions**: Each session has:
  - Chromium browser instance
  - HTTP server on unique port (starting from 8080)
  - Steam login automation
- **Control Server**: Master API on port 9000

## Installation

### 1. Build the binary

```bash
go build -o csrep
```

### 2. Install Playwright browsers

```bash
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install chromium
```

## Usage

### Start the server

```bash
./csrep server
```

This starts:
- Control API on port 9000
- Interactive CLI for managing sessions
- Sessions start from port 8080 onwards

### Add a Session

```bash
> add myusername
```

This will:
1. Auto-generate a session ID (e.g., `session-1`)
2. Open a Chrome browser window
3. Navigate to Steam login page
4. Wait for you to manually log in

**After logging in manually**, mark the login as complete:

```bash
curl -X POST http://localhost:8080/action/login-complete
```

### Interactive Commands

```bash
# Add a new session
add username123

# List all active sessions
list

# Remove a session
remove session-1

# Exit
quit
```

## API Endpoints

### Control API (Port 9000)

**List all sessions:**
```bash
curl http://localhost:9000/sessions
```

**Add a session:**
```bash
curl -X POST http://localhost:9000/sessions/add \
  -H "Content-Type: application/json" \
  -d '{"userId":"username123"}'
```

### Session APIs (Individual Ports)

Each session gets its own port starting from 8080.

**Get session status:**
```bash
curl http://localhost:8080/status
```

**Mark login as complete:**
```bash
curl -X POST http://localhost:8080/action/login-complete
```

**Get screenshot:**
```bash
curl http://localhost:8080/action/get-screenshot > screenshot.png
```

**Navigate to URL:**
```bash
curl -X POST http://localhost:8080/action/navigate \
  -H "Content-Type: application/json" \
  -d '{"url":"https://steamcommunity.com/market"}'
```

**Open Steam profile:**
```bash
curl -X POST http://localhost:8080/action/open-profile
```

## Example Workflow

```bash
# 1. Start the server
./csrep server

# 2. Add a session (browser opens automatically)
> add trader_bot

# Output:
# ✓ Session 'session-1' created on port 8080
#   Browser window opened at Steam login page
#   Please log in manually in the browser
#   After login, mark complete: curl -X POST http://localhost:8080/action/login-complete

# 3. Complete login in the browser window manually

# 4. In another terminal, mark login complete
curl -X POST http://localhost:8080/action/login-complete

# 5. Add more sessions
> add trader_bot2
> add trader_bot3

# 6. List all sessions
> list

# SESSION ID      USERNAME             PORT       LOGGED IN   
# --------------------------------------------------------------
# session-1       trader_bot           8080       Yes         
# session-2       trader_bot2          8081       No          
# session-3       trader_bot3          8082       No          

# 7. Trigger actions
curl -X POST http://localhost:8080/action/navigate -d '{"url":"https://steamcommunity.com/market"}'
```

## Login Flow

1. **Automated**: Browser opens and navigates to Steam login
2. **Manual**: You enter credentials and complete 2FA/Steam Guard
3. **Hybrid**: Call `/action/login-complete` to mark session as ready
4. **Automated**: Bot can now execute actions on your behalf

## Requirements

- Go 1.22+
- Playwright for Go
- Chromium (installed via Playwright)
- Linux with X11/Wayland display
