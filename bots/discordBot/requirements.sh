#!/usr/bin/env bash
# install_prereqs.sh - Install prerequisites for csrep project
# Usage: ./install_prereqs.sh

set -e

# Install Go (if not present)
if ! command -v go &> /dev/null; then
  echo "Go not found. Please install Go 1.22+ from https://go.dev/dl/ and re-run this script."
  exit 1
fi

echo "Go version: $(go version)"

# Install Playwright for Go (Go module)
echo "Installing Playwright Go module..."
go get github.com/playwright-community/playwright-go

# Install Playwright browsers (Chromium)
echo "Installing Playwright browsers (Chromium)..."
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install chromium

# Linux dependencies for Chromium
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
  echo "Installing Linux dependencies for Chromium..."
  sudo apt update && sudo apt install -y libnss3 libatk-bridge2.0-0 libgtk-3-0 libxss1 libasound2 libgbm1
fi

echo "All prerequisites installed!"
