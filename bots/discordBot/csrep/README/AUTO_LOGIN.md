# Steam Account Auto-Login

## Overview

The bot now supports **automatic Steam login** using pre-stored credentials. Browsers will automatically fill in login details without any manual input required.

## How It Works

1. **Store Credentials First** - Save Steam account credentials
2. **Create Session** - Bot automatically logs in with saved credentials
3. **Manual Fallback** - If no credentials or auto-login fails, falls back to manual login

## Credential Management

### Add Credentials

```bash
./csrep server

> cred-add myaccount mypassword
✓ Credentials saved for account: myaccount

# With Steam Guard 2FA shared secret
> cred-add myaccount mypassword ABCDEF123456
✓ Credentials saved for account: myaccount
  (includes Steam Guard shared secret)
```

### List Stored Accounts

```bash
> cred-list

Stored Steam Accounts:
  - myaccount
  - botaccount1
  - botaccount2
```

### Remove Credentials

```bash
> cred-remove myaccount
✓ Credentials removed for: myaccount
```

## Creating Sessions with Auto-Login

Once credentials are stored, sessions automatically log in:

```bash
# Add session for account with stored credentials
> add myaccount

Creating session session-1 for user myaccount on port 8080
Found credentials for myaccount, will auto-login
Browser will open and auto-login with saved credentials...
[Session session-1] Auto-logging in with account: myaccount
[Session session-1] Credentials entered, checking for 2FA...
[Session session-1] ✓ Login detected! Starting API server...
✓ Session session-1 is now fully logged in and ready!
```

**Without credentials:**
```bash
> add newaccount

Creating session session-2 for user newaccount on port 8081
No credentials found for newaccount, manual login required
Browser will open - please complete Steam login...
# (waits for manual login)
```

## Credential Storage

### File Location
- Stored in: `steam_credentials.json`
- File permissions: `0600` (owner read/write only)
- **Auto-gitignored** - won't be committed

### Format
```json
{
  "myaccount": {
    "account_name": "myaccount",
    "password": "mypassword",
    "shared_secret": "ABCDEF123456"
  }
}
```

## Security Features

✅ **File Permissions** - Only owner can read/write credentials  
✅ **Gitignored** - Never committed to version control  
✅ **Encrypted Storage** - TODO: Add encryption at rest  
✅ **2FA Support** - Shared secret for Steam Guard TOTP

## Steam Guard 2FA

### Getting Your Shared Secret

If you use Steam Mobile Authenticator, you need the shared secret:

1. **Android:**
   - Root required
   - Location: `/data/data/com.valvesoftware.android.steam.community/files/Steamguard-<SteamID>`
   - Extract `shared_secret` from JSON

2. **Using SteamDesktopAuthenticator:**
   - Download: https://github.com/Jessecar96/SteamDesktopAuthenticator
   - Extract `shared_secret` from maFile

3. **Store with credentials:**
   ```bash
   > cred-add myaccount mypassword YOUR_SHARED_SECRET
   ```

### Current 2FA Status

⚠️ **Steam Guard TOTP generation not yet implemented**

Currently, if 2FA is required:
- Bot will enter username/password
- Bot will detect 2FA prompt
- **You must manually enter the code** from your mobile app
- Future: Auto-generate code from shared secret

## Workflow Examples

### Scenario 1: First Time Setup

```bash
# 1. Add credentials
> cred-add trader_bot mypassword123

# 2. Create session (auto-logins)
> add trader_bot
✓ Session 'session-1' is READY on port 8080
```

### Scenario 2: Multiple Accounts

```bash
# Store multiple accounts
> cred-add bot1 pass1
> cred-add bot2 pass2
> cred-add bot3 pass3

# List them
> cred-list
  - bot1
  - bot2
  - bot3

# Create sessions (all auto-login)
> add bot1
> add bot2
> add bot3
```

### Scenario 3: Mixed Manual/Auto

```bash
# Some with credentials
> add bot1  # auto-logins
> add bot2  # auto-logins

# Some without (manual login)
> add newbot  # opens browser, you log in manually
```

## Command Reference

```bash
# Credential Management
cred-add <account> <password> [shared_secret]  - Add credentials
cred-list                                      - List all accounts
cred-remove <account>                          - Remove credentials

# Session Creation (unchanged)
add <username> [timeout]  - Create session (auto if creds exist)
list                      - List sessions
remove <session-id>       - Remove session
```

## Security Best Practices

1. **Protect credentials file:**
   ```bash
   chmod 600 steam_credentials.json
   ```

2. **Never share credentials file**

3. **Use separate accounts for bots** - Don't use your main Steam account

4. **Rotate passwords regularly**

5. **Enable Steam Guard** - Always use 2FA on bot accounts

## Troubleshooting

**Auto-login not working:**
- Check credentials: `cred-list`
- Verify account name matches exactly
- Check browser console for errors
- May need manual 2FA code entry

**2FA issues:**
- Shared secret must be correct base32 string
- TOTP generation coming in future update
- For now, enter codes manually

**Credentials not saving:**
- Check file permissions
- Verify disk space
- Check for JSON syntax errors

## Future Enhancements

- [ ] Steam Guard TOTP code generation
- [ ] Encrypted credential storage
- [ ] Import/export credentials
- [ ] Credential testing before storage
- [ ] Session cookies for faster re-login
