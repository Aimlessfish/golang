# Proxy Handler

This Go package provides functionality to retrieve proxy servers by executing a one-shot binary that scrapes proxy data from API providers.

## Overview

The `ProxyHandler` function executes a binary located at `./bin/proxy` with a specified proxy type argument (`http`, `https`, or `socks5`). The binary outputs JSON data containing proxy information, which is then parsed and returned as a slice of strings in the format `ip:port`.

## Binary Execution

The binary is executed as a one-shot process using `util.ExecBinary()`. It takes one argument specifying the proxy type:

- `http` - Retrieves HTTP proxies
- `https` - Retrieves HTTPS proxies
- `socks5` - Retrieves SOCKS5 proxies

## Output Format

The binary outputs a JSON array of proxy objects:

```json
[
  {
    "ip": "154.3.236.202",
    "port": "3128"
  },
  {
    "ip": "101.47.16.15",
    "port": "7890"
  }
]
```

The function parses this JSON and converts each proxy object into a string: `"ip:port"`.

## Usage

```go
proxies := proxy.ProxyHandler("http")
// Returns: ["154.3.236.202:3128", "101.47.16.15:7890", ...]
```

## Error Handling

- If the binary execution fails, an empty slice is returned and an error is logged.
- If JSON parsing fails, an empty slice is returned and an error is logged.
- Invalid proxy types result in an empty slice and an error log.

## Dependencies

- `discordBot/util` - For binary execution and logging utilities
- `encoding/json` - For JSON parsing

## Binary Requirements

The `./bin/proxy` binary must be present and executable. It should accept one argument (proxy type) and output valid JSON to stdout.
