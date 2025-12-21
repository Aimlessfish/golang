# Discord Bot Usage Guide

## Overview

This bot is designed for **fully automated Steam workflows** via Discord. The only manual steps are:
1. **Adding bot accounts** to the credential pool
2. **Issuing commands** via Discord

Everything else (browser automation, login, navigation) is handled automatically.

## Quick Start

### 1. Setup

```bash
# Create .env file
cp .env.example .env

# Edit .env and add your Discord bot token
DISCORD_BOT_TOKEN=your_actual_token_here
```

### 2. Start Discord Bot

```bash
./csrep discord
```

### 3. Add Bot Accounts (via Discord)

In any Discord channel where the bot is present:

```
!bot-add email@example.com botusername botpassword
!bot-add email@example.com botusername botpassword ABC123SHAREDSECRET
```

The bot will:
- ✅ Store credentials securely
- ✅ Delete your message for security
- ✅ Confirm the account was added

### 4. Use Automation (via Discord)

```
!report https://steamcommunity.com/id/targetuser
```

The bot will:
- ✅ Select an available bot account
- ✅ Create a browser session
- ✅ Auto-login with credentials
- ✅ Navigate to the target profile
- ✅ Execute report automation (when implemented)

## Discord Commands

### Bot Management

**Add bot account:**
```
!bot-add <email> <username> <password> [shared_secret]
```
Example:
```
!bot-add bot1@gmail.com steambot1 mypassword123
!bot-add bot2@gmail.com steambot2 mypassword456 ABCDEF123456
```

**List bot accounts:**
```
!bot-list
```
Shows all available bot accounts in the pool.

**Remove bot account:**
```
!bot-remove <username>
```
Example:
```
!bot-remove steambot1
```

### Actions

**Report a player:**
```
!report <steam_profile_url>
```
Example:
```
!report https://steamcommunity.com/id/cheater123
!report https://steamcommunity.com/profiles/76561198012345678
```

**View active sessions:**
```
!sessions
```
Shows all browser sessions currently running.

**Help:**
```
!help
```
Shows command list.

## Workflow Examples

### Example 1: Report Player

```
User: !report https://steamcommunity.com/id/cheater
Bot:  🤖 Starting report process...
      📝 Using bot: steambot1
      🎯 Target: https://steamcommunity.com/id/cheater

Bot:  ✅ Session session-1 created and logged in on port 8080
      ✅ Navigated to profile
      
      [Report automation executes automatically]
```

### Example 2: Add Multiple Bots

```
User: !bot-add bot1@gmail.com bot1 pass1 SECRET1
Bot:  ✅ Bot account added: bot1
      📧 Email: bot1@gmail.com
      🔐 2FA: Enabled
      (Original message deleted for security)

User: !bot-add bot2@gmail.com bot2 pass2
Bot:  ✅ Bot account added: bot2
      📧 Email: bot2@gmail.com
      (Original message deleted for security)

User: !bot-list
Bot:  🤖 Available Bot Accounts:
      1. bot1
      2. bot2
```

### Example 3: Monitor Sessions

```
User: !sessions
Bot:  🌐 Active Sessions:
      ID           Bot Account     Port   Status
      ──────────────────────────────────────────
      session-1    bot1           8080   Ready
      session-2    bot2           8081   Ready
```

## Architecture

### Fully Automated Flow

```
Discord Command (!report <url>)
    ↓
Select Available Bot Account
    ↓
Create Browser Session
    ↓
Auto-Login (credentials from pool)
    ↓
Navigate to Target Profile
    ↓
Execute Report Automation
    ↓
Clean Up Session
    ↓
Report Result to Discord
```

### Zero Manual Intervention

Once bot accounts are added:
- ✅ No manual browser interaction needed
- ✅ No manual login needed
- ✅ No manual navigation needed
- ✅ Everything triggered via Discord commands

## Security

### Credential Protection

- Credentials stored in `steam_credentials.json` (file permissions: 0600)
- File is **gitignored** - never committed
- Discord messages with credentials are **auto-deleted**
- Only username shown in bot responses

### Best Practices

1. **Use dedicated bot accounts** - Never use personal Steam accounts
2. **Separate emails** - One email per bot account
3. **Limited value** - Keep minimal items/wallet balance on bot accounts
4. **Monitor usage** - Check bot activity regularly
5. **Rotate credentials** - Change passwords periodically

## Adding Report Automation Logic

The framework is ready - you can add custom automation later. Example:

```go
// In handleReport function, after navigation:

// Click report button
if err := session.ClickElement("button.report_button"); err != nil {
    return err
}

// Wait for reason dropdown
if err := session.WaitForSelector("select.report_reason", 5000); err != nil {
    return err
}

// Select reason (e.g., "Cheating")
page := session.GetPage()
page.SelectOption("select.report_reason", playwright.SelectOptionValues{
    Values: &[]string{"cheating"},
})

// Submit report
if err := session.ClickElement("button.submit_report"); err != nil {
    return err
}

// Wait for confirmation
session.WaitForSelector(".success_message", 5000)
```

Helper methods available:
- `session.ClickElement(selector)` - Click element
- `session.WaitForSelector(selector, timeout)` - Wait for element
- `session.GetPage()` - Get Playwright page for advanced automation
- `session.NavigateTo(url)` - Navigate to URL
- `session.GetScreenshot()` - Take screenshot

## Troubleshooting

**Bot not responding:**
- Check `DISCORD_BOT_TOKEN` in `.env`
- Verify bot has permission to read messages in channel
- Check bot is online in Discord

**Auto-login failing:**
- Verify credentials are correct: `!bot-list`
- Check Steam Guard 2FA shared secret (if enabled)
- Check browser console for errors

**No bot accounts available:**
- Add accounts: `!bot-add email username password`
- Check credentials exist: `!bot-list`

## Development

### CLI Mode (for testing)

You can also run in CLI mode for testing:

```bash
./csrep server

# Add credentials
> cred-add email@test.com botname password

# Manually create session
> add botname
```

### Discord Mode (production)

```bash
./csrep discord
```

All interaction happens in Discord - no CLI needed.

## Next Steps

1. ✅ Add bot accounts to credential pool
2. ✅ Test with `!report` command
3. ⏳ Implement custom report automation logic
4. ⏳ Add more automation commands (e.g., `!trade`, `!message`, etc.)

## Command Reference

| Command | Arguments | Description |
|---------|-----------|-------------|
| `!bot-add` | email, username, password, [2fa] | Add bot account |
| `!bot-list` | - | List bot accounts |
| `!bot-remove` | username | Remove bot account |
| `!report` | steam_url | Report player (automated) |
| `!sessions` | - | View active sessions |
| `!help` | - | Show commands |
