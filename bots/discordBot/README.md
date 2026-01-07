# Discord Steam Market Bot

This bot provides Discord commands for interacting with the Steam Market and other utilities.

## Features
- **!market <item_name>**: Fetches Steam Market data for the specified item. Example: `!market AK-47 | Nightwish (Field-Tested)`
- **!report <uid> <amount>**: Starts a Steam report for a user ID.
- **!bot-add <username> <password>**: Adds a bot account.
- **!bot-remove <username>**: Removes a bot account.
- **!bot-list**: Lists all bot accounts.
- **!proxy**: Returns a proxy.
- **!clear**: Clears bot messages.
- **!help**: Lists available commands.

## Usage
1. Clone the repository and install Go dependencies:
   ```fish
   git clone <repo-url>
   cd discordBot
   go mod tidy
   ```
2. Set up your `.env` file or environment variables for Discord and Steam credentials.
3. Run the bot:
   ```fish
   go run main.go
   ```

## Steam Market Command Example
```
!market AK-47 | Nightwish (Field-Tested)
```
This will return the current market data for the specified item.

## Requirements
- Go 1.18+
- Discord bot token
- Steam credentials (see `steam_credentials.json`)

## Project Structure
- `main.go` - Entry point
- `bot/` - Discord bot logic
- `functions/` - Command handlers
- `util/` - Utility functions

## License
MIT
