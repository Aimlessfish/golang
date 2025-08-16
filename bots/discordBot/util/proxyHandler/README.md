# Free Proxy API Scraper and Tester
Author: Will Meekin

## Description:
This Go program scrapes free HTTP proxies from two public APIs and returns as many working proxies as possible. It supports returning either a list of proxies or a single proxy depending on the selected mode.


## Features
Scrapes proxy lists from:
    PubProxy
    ProxyScrape
Validates proxies by attempting to open a TCP connection to google.com:80.
Returns either:
    A list of working proxies
    A single working proxy (first found)

## Usage
-mode=0	Return a list of all working proxies
-mode=1	Return only a single working proxy
 Example:
    ```go run /main.go -mode 1 ``` to return 1 working HTTP proxy.

## Structure
proxyHandler/
│
├── apiCalls/
│   └── apiCalls.go         # Functions to fetch proxies from both APIs
├── main.go                 # Entry point: handles flags, calls APIs, tests proxies 