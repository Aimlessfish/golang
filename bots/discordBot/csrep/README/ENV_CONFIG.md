# Environment Configuration Guide

## Overview

The bot now uses `.env` files to manage secrets and configuration. This keeps sensitive data out of your code and makes deployment easier.

## Setup

1. **Copy the example file:**
   ```bash
   cp .env.example .env
   ```

2. **Edit `.env` with your values:**
   ```bash
   nano .env  # or your preferred editor
   ```

3. **Add your Discord bot token:**
   ```env
   DISCORD_BOT_TOKEN=your_actual_token_here
   ```

## Available Configuration

### Required (for Discord mode)
- `DISCORD_BOT_TOKEN` - Your Discord bot token from Discord Developer Portal

### Optional
- `SESSION_START_PORT=8080` - Starting port for bot sessions (default: 8080)
- `SESSION_TIMEOUT_MINUTES=30` - Default auto-remove timeout (default: 0/disabled)
- `CONTROL_API_PORT=9000` - Control API port (default: 9000)
- `HEADLESS_BROWSER=false` - Run browsers without GUI (default: false)

## Usage Examples

### CLI Mode
```bash
# Uses .env for defaults
./csrep server

# Sessions will:
# - Start on port from SESSION_START_PORT
# - Use SESSION_TIMEOUT_MINUTES as default timeout
# - Can override timeout per session: add username 60
```

### Discord Mode
```bash
# Requires DISCORD_BOT_TOKEN in .env
./csrep discord
```

If token is missing, you'll see:
```
❌ Error: DISCORD_BOT_TOKEN not set

Please create a .env file with your Discord bot token:
  DISCORD_BOT_TOKEN=your_token_here
```

## Security Best Practices

1. **Never commit `.env` to git**
   - Already in `.gitignore`
   - Only commit `.env.example`

2. **Use different tokens for dev/prod**
   - Keep separate `.env` files
   - Use `.env.production`, `.env.development`

3. **Rotate tokens regularly**
   - Update `.env` file
   - Restart the bot

4. **Permissions**
   ```bash
   chmod 600 .env  # Only you can read/write
   ```

## Getting a Discord Bot Token

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click "New Application"
3. Go to "Bot" section
4. Click "Reset Token" and copy the token
5. Paste into `.env` as `DISCORD_BOT_TOKEN=...`

## Environment Variables Priority

1. System environment variables (highest)
2. `.env` file
3. Default values in code (lowest)

Override any `.env` value:
```bash
SESSION_START_PORT=9090 ./csrep server
```

## Troubleshooting

**Bot won't start in Discord mode:**
- Check `.env` exists
- Verify `DISCORD_BOT_TOKEN` is set
- Token should be ~70 characters
- No quotes around token value

**Ports already in use:**
- Change `SESSION_START_PORT` in `.env`
- Or use different value: `SESSION_START_PORT=9000 ./csrep server`

**Can't find .env:**
- Must be in same directory as binary
- Or set full path: `DOTENV_PATH=/path/to/.env ./csrep server`
