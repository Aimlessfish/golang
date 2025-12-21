#!/bin/bash

# Steam user profile opener
# Takes a Steam user ID as an argument and opens it in the default browser

if [ -z "$1" ]; then
    echo "Error: Steam user ID is required"
    exit 1
fi

STEAM_USER_ID="$1"
STEAM_URL="https://steamcommunity.com/id/${STEAM_USER_ID}"

echo "Opening Steam profile: ${STEAM_URL}"

# Detect the platform and use appropriate command to open URL
if command -v xdg-open &> /dev/null; then
    # Linux
    xdg-open "${STEAM_URL}"
elif command -v open &> /dev/null; then
    # macOS
    open "${STEAM_URL}"
elif command -v start &> /dev/null; then
    # Windows (Git Bash, WSL, etc.)
    start "${STEAM_URL}"
else
    echo "Error: Could not detect a command to open URLs"
    echo "Please open manually: ${STEAM_URL}"
    exit 1
fi
