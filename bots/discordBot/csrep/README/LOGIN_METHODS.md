# Steam Login Methods

The bot now supports **two authentication methods** for Steam login:

1. **Username + Password** - Traditional login with credentials
2. **QR Code** - Scan QR code with Steam Mobile app

## Quick Comparison

| Method | Setup | Security | 2FA Required | Best For |
|--------|-------|----------|--------------|----------|
| **Password** | Store username + password | Credentials in file | Optional | Automation, multiple accounts |
| **QR Code** | Just account name | No password stored | Required (mobile app) | Single account, better security |

## Method 1: Username + Password

### Setup

```bash
./csrep server

# Without 2FA
> cred-add myaccount mypassword

# With Steam Guard 2FA
> cred-add myaccount mypassword SHARED_SECRET_HERE
```

### How It Works

1. Bot opens Steam login page
2. **Automatically fills** username and password
3. **Automatically clicks** sign in button
4. If 2FA enabled and shared secret provided, attempts to enter code
5. Session becomes active

### Advantages

✅ **Fully automated** - no manual interaction needed  
✅ **Multiple accounts** - can manage many bot accounts  
✅ **Fast login** - instant form filling  
✅ **Optional 2FA** - can store shared secret for auto-2FA  

### Disadvantages

⚠️ **Security risk** - password stored in plaintext JSON  
⚠️ **Steam Guard complexity** - TOTP generation needs implementation  
⚠️ **Account security** - violates Steam ToS if password is shared  

## Method 2: QR Code

### Setup

```bash
./csrep server

# Set up QR login (no password needed)
> cred-add myaccount qr
```

### How It Works

1. Bot opens Steam QR login page
2. **You scan the QR code** with your Steam Mobile app
3. Confirm login on your phone
4. Session becomes active automatically

### Advantages

✅ **More secure** - no password stored anywhere  
✅ **Official method** - supported by Steam  
✅ **Easy setup** - just account name needed  
✅ **Steam Guard built-in** - uses mobile authenticator  

### Disadvantages

⚠️ **Manual interaction** - must scan QR each time  
⚠️ **Requires phone** - need Steam Mobile app  
⚠️ **Not fully automated** - user must scan code  

## Usage Examples

### Example 1: Bot Account with Password

For automated bot accounts that run unattended:

```bash
# Store credentials with password
> cred-add trader_bot secretpass123 ABC123SHAREDSECRET

# Create session (auto-logins)
> add trader_bot
✓ Session 'session-1' is READY on port 8080
```

**Best for:** Trading bots, market bots, automated tasks

### Example 2: Personal Account with QR

For your main Steam account with better security:

```bash
# Store account for QR login
> cred-add mysteamname qr

# Create session (opens QR code)
> add mysteamname
# [Browser opens with QR code]
# [Scan with your phone]
✓ Session 'session-1' is READY on port 8080
```

**Best for:** Personal accounts, testing, one-off tasks

### Example 3: Mixed Accounts

Different methods for different accounts:

```bash
# Bot accounts with password
> cred-add bot1 pass1
> cred-add bot2 pass2
> cred-add bot3 pass3

# Personal account with QR
> cred-add mainaccount qr

# All can be used the same way
> add bot1      # auto-logins with password
> add bot2      # auto-logins with password
> add mainaccount  # shows QR code to scan
```

## Checking Your Credentials

```bash
> cred-list

Stored Steam Accounts:
  - bot1 (password)
  - bot2 (password)
  - mainaccount (QR code)
```

## Switching Login Methods

To change an account's login method, remove and re-add:

```bash
# Switch from password to QR
> cred-remove myaccount
> cred-add myaccount qr

# Switch from QR to password
> cred-remove myaccount
> cred-add myaccount newpassword
```

## Security Recommendations

### For Bot Accounts (Password Method)

1. **Use dedicated accounts** - never use your main Steam account
2. **Separate email** - use unique email per bot account
3. **Limited value** - keep minimal items/wallet balance
4. **Rotate passwords** - change passwords regularly
5. **Monitor activity** - check for unauthorized access

### For Personal Accounts (QR Method)

1. **Always use QR** - better security, official Steam method
2. **Don't store passwords** - even if more convenient
3. **Keep phone secure** - protect your Steam Mobile app
4. **Enable all security** - use all Steam Guard features
5. **Review sessions** - check logged-in devices regularly

## Technical Details

### Password Login Flow

1. Navigate to `https://store.steampowered.com/login/`
2. Query selector: `input[type='text']` for username
3. Query selector: `input[type='password']` for password
4. Query selector: `button[type='submit']` for sign in
5. Auto-fill username → password → click
6. Detect 2FA prompt if present
7. Monitor URL change for successful login

### QR Login Flow

1. Navigate to `https://steamcommunity.com/login/home/?goto=`
2. QR code automatically displayed by Steam
3. User scans with Steam Mobile app
4. User confirms on phone
5. Steam automatically redirects on success
6. Monitor URL change for successful login

### Credential Storage Format

```json
{
  "myaccount": {
    "account_name": "myaccount",
    "password": "mypassword",
    "shared_secret": "ABC123",
    "login_method": "password"
  },
  "qraccount": {
    "account_name": "qraccount",
    "login_method": "qr"
  }
}
```

## Troubleshooting

### Password login not working

- Verify credentials: `cred-list`
- Check for typos in username/password
- Check if Steam requires CAPTCHA (not supported yet)
- Try manual login to verify credentials work

### QR code not showing

- Check you're using `qr` as the second parameter
- Verify Steam's QR login page loads
- Try refreshing the browser page
- Check Steam service status

### 2FA issues with password login

- TOTP generation not yet implemented
- Must enter codes manually for now
- Or use QR method which has built-in 2FA

## Future Enhancements

- [ ] CAPTCHA handling for password login
- [ ] Steam Guard TOTP code generation
- [ ] Automatic QR detection and confirmation
- [ ] Session cookies for faster re-login
- [ ] Encrypted password storage
- [ ] Mobile notifications for QR scans

## Command Reference

```bash
# Password login
cred-add <account> <password> [shared_secret]

# QR login
cred-add <account> qr

# List all credentials (shows login method)
cred-list

# Remove credentials
cred-remove <account>

# Create session (uses stored method)
add <account>
```
