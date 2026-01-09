#!/bin/bash

# One-Shot Installer for Telegram Shop Bot
# This script installs dependencies and sets up the bot

set -e  # Exit on any error

echo "Starting one-shot installer for Telegram Shop Bot..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Installing Go..."
    # Assuming Ubuntu/Debian, adjust for other distros
    sudo apt update
    sudo apt install -y golang-go
    echo "Go installed successfully."
else
    echo "Go is already installed."
fi

# Verify Go version (optional)
go version

# Navigate to the project directory
cd "$(dirname "$0")"

# Download Go dependencies
echo "Downloading Go dependencies..."
go mod download

# Tidy up go.mod
go mod tidy

# Build the project
echo "Building the project..."
go build -o tgShopBot .

# Check for .env file and prompt for bot token
if [ ! -f .env ]; then
    echo "Creating .env file..."
    echo -n "Enter your Telegram Bot Token (get from @BotFather): "
    read -s bot_token
    echo ""  # New line after hidden input
    cat > .env << EOF
# Telegram Bot Token
TELEGRAM_BOT_TOKEN=$bot_token

# Database configuration (placeholders for future use)
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=telegram_shop
EOF
    echo ".env file created with your bot token."
else
    echo ".env file already exists."
    echo -n "Enter your Telegram Bot Token to update (or press Enter to skip): "
    read -s bot_token
    echo ""  # New line
    if [ -n "$bot_token" ]; then
        # Update the token in existing .env
        sed -i "s/^TELEGRAM_BOT_TOKEN=.*/TELEGRAM_BOT_TOKEN=$bot_token/" .env
        echo "Bot token updated in .env."
    else
        echo "Bot token not updated."
    fi
fi

echo "Installation complete! Run './tgShopBot' to start the bot."
echo "Make sure to set your TELEGRAM_BOT_TOKEN in the .env file."

# Database setup section
echo ""
echo "Would you like to set up a local PostgreSQL database with the shop schema? (y/n)"
read -r setup_db
if [[ $setup_db =~ ^[Yy]$ ]]; then
    echo "Setting up PostgreSQL database..."

    # Check if PostgreSQL is installed
    if ! command -v psql &> /dev/null; then
        echo "Installing PostgreSQL..."
        sudo apt update
        sudo apt install -y postgresql postgresql-contrib
        sudo systemctl start postgresql
        sudo systemctl enable postgresql
        echo "PostgreSQL installed and started."
    else
        echo "PostgreSQL is already installed."
    fi

    # Create database and user
    echo "Creating database and user..."
    sudo -u postgres psql -c "CREATE USER telegram_bot WITH PASSWORD 'bot_password';" 2>/dev/null || echo "User telegram_bot already exists."
    sudo -u postgres psql -c "CREATE DATABASE telegram_shop OWNER telegram_bot;" 2>/dev/null || echo "Database telegram_shop already exists."

    # Update .env with database credentials
    sed -i "s/DB_HOST=.*/DB_HOST=localhost/" .env
    sed -i "s/DB_PORT=.*/DB_PORT=5432/" .env
    sed -i "s/DB_USER=.*/DB_USER=telegram_bot/" .env
    sed -i "s/DB_PASSWORD=.*/DB_PASSWORD=bot_password/" .env
    sed -i "s/DB_NAME=.*/DB_NAME=telegram_shop/" .env

    # Run the schema
    echo "Running database schema..."
    PGPASSWORD=bot_password psql -h localhost -U telegram_bot -d telegram_shop -f schema.sql

    echo "Database setup complete!"
    echo "Database credentials have been added to .env"
else
    echo "Skipping database setup. You can manually run the schema.sql file on your database."
fi

# Proxy chain setup section
echo ""
echo "Would you like to set up proxy chaining for the bot? (y/n)"
read -r setup_proxy
if [[ $setup_proxy =~ ^[Yy]$ ]]; then
    echo "Setting up proxy chains..."

    # Install proxychains
    if ! command -v proxychains &> /dev/null; then
        echo "Installing proxychains..."
        sudo apt update
        sudo apt install -y proxychains
        echo "proxychains installed."
    else
        echo "proxychains is already installed."
    fi

    # Prompt for proxies
    echo "Enter your HTTPS proxies (one per line, format: ip:port or user:pass@ip:port)"
    echo "Press Ctrl+D when done:"
    proxy_list=""
    while IFS= read -r proxy; do
        if [ -n "$proxy" ]; then
            proxy_list="$proxy_list$proxy\n"
        fi
    done

    # Configure proxychains
    sudo cp /etc/proxychains.conf /etc/proxychains.conf.backup
    sudo bash -c "cat > /etc/proxychains.conf << EOF
strict_chain
proxy_dns
remote_dns_subnet 224
tcp_read_time_out 15000
tcp_connect_time_out 8000
[ProxyList]
EOF"

    # Add user proxies
    if [ -n "$proxy_list" ]; then
        echo -e "$proxy_list" | sudo tee -a /etc/proxychains.conf > /dev/null
    else
        # Add some default example proxies
        sudo bash -c "cat >> /etc/proxychains.conf << EOF
# Example proxies - replace with your own
# http 192.168.1.1 8080
# socks5 192.168.1.2 1080 user pass
EOF"
    fi

    echo "Proxy chains configured."
    echo "To run the bot with proxy chains, use: proxychains ./tgShopBot"
    echo "Note: Make sure your proxies are working and allow HTTPS traffic."
else
    echo "Skipping proxy setup."
fi