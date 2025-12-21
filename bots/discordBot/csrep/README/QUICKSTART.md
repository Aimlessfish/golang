# Steam Bot - Quick Start

## What Changed

The project has been restructured for **Discord bot integration** with a **modular feature system**:

### New Structure

```
bot/              # Core library (reusable)
├── session.go    # Browser session management
├── manager.go    # Multi-session orchestration + timeouts
├── features.go   # Feature interface
└── features_impl.go  # Built-in features

cmd/              # Applications
├── main.go       # CLI + Discord entry points
└── discord.go    # Discord command handlers
```

### Key Features Added

✅ **Session Auto-Removal with Timeouts**
- Sessions can auto-expire after X minutes
- Extend timeout on-demand
- Perfect for Discord rate limiting

✅ **Modular Feature System**
- Navigate, Screenshot, Status, Open Profile
- Easy to add custom features
- Feature interface for consistency

✅ **Discord-Ready Framework**
- Command handlers pre-built
- Session management via Discord
- Screenshot upload support ready

✅ **Blocking Login Flow**
- Sessions don't accept commands until logged in
- Auto-detects successful Steam login
- No manual "login complete" needed

## Building

```bash
# Build new version
go build -o csrep ./cmd/

# Old files (main.go, manager.go, session.go in root) are now legacy
```

## CLI Usage

```bash
# Start CLI server
./csrep server

# Commands in CLI:
> add username          # Create session (no timeout)
> add username 30       # Create session, auto-remove after 30min
> list                  # Show all sessions
> extend session-1 15   # Add 15 more minutes
> remove session-1      # Manual removal
> quit                  # Stop all
```

## Discord Mode (Framework Ready)

```bash
./csrep discord
```

The Discord integration structure is complete. To finish:

1. Add `github.com/bwmarrin/discordgo` to dependencies
2. Wire up message handlers in `cmd/discord.go`
3. Set Discord bot token
4. Deploy!

**Planned Commands:**
- `!steam add <username> [timeout]`
- `!steam remove <session-id>`
- `!steam list`
- `!steam screenshot <session-id>` (uploads PNG)
- `!steam navigate <session-id> <url>`

## Feature System

Create custom features:

```go
type MyFeature struct{}

func (f *MyFeature) Name() string {
    return "my-feature"
}

func (f *MyFeature) Execute(session *bot.Session, args map[string]interface{}) (interface{}, error) {
    // Your automation logic here
    return result, nil
}
```

## Timeout Management

```go
// Add session with 30 minute auto-remove
session, _ := manager.AddSessionWithTimeout("user", 30*time.Minute)

// Extend by 15 more minutes
manager.ExtendTimeout("session-1", 15*time.Minute)

// Remove manually
manager.RemoveSession("session-1")
```

## Migration Notes

- Old `main.go`, `manager.go`, `session.go` in root still exist but are legacy
- New code uses `bot/` package
- CLI works exactly the same but with timeout support
- Discord mode is new

See `DISCORD.md` for full Discord integration guide.
