# Example: Running 3 Bot Sessions

# Terminal 1: Start the server
./csrep server

# The server will show:
# Steam Bot Session Manager started
# Control API running on :9000
# Sessions will start from port 8080
# >

# Add three sessions interactively
> add trader1 steam_user_001
# Session 'trader1' created on port 8080

> add trader2 steam_user_002
# Session 'trader2' created on port 8081

> add trader3 steam_user_003
# Session 'trader3' created on port 8082

> list
# SESSION ID      USER ID              PORT      
# --------------------------------------------------
# trader1         steam_user_001       8080      
# trader2         steam_user_002       8081      
# trader3         steam_user_003       8082      

# Terminal 2: Trigger actions on the sessions
# Check health of all sessions
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health

# Open Steam profiles for each bot
curl -X POST http://localhost:8080/action/open-profile
curl -X POST http://localhost:8081/action/open-profile
curl -X POST http://localhost:8082/action/open-profile

# Navigate bots to different URLs
curl -X POST http://localhost:8080/action/navigate \
  -H "Content-Type: application/json" \
  -d '{"url":"https://steamcommunity.com/market"}'

curl -X POST http://localhost:8081/action/navigate \
  -H "Content-Type: application/json" \
  -d '{"url":"https://steamcommunity.com/market/listings/730/AK-47"}'

curl -X POST http://localhost:8082/action/navigate \
  -H "Content-Type: application/json" \
  -d '{"url":"https://steamcommunity.com/profiles/steam_user_003/inventory"}'

# Terminal 3: Use Control API
# List all sessions programmatically
curl http://localhost:9000/sessions | jq

# Add a new session via API
curl -X POST http://localhost:9000/sessions/add \
  -H "Content-Type: application/json" \
  -d '{"id":"trader4","userId":"steam_user_004"}'

# Back to Terminal 1: Stop everything
> quit
