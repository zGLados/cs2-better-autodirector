# 🧪 FACEIT Integration Testing

Quick test script to verify FACEIT API integration without starting the full GUI.

**Location:** `scripts/test_faceit.go`

## Usage

### Method 1: Using secrets.json (Recommended)

1. Make sure `config/secrets.json` contains your API key:
   ```json
   {
     "faceit_api_key": "your-api-key-here"
   }
   ```

2. Run the test with a FACEIT match URL:
   ```bash
   cd scripts
   go run test_faceit.go -url "https://www.faceit.com/en/cs2/room/1-abc123-def456-..."
   ```

### Method 2: Provide API key via command line

```bash
cd scripts
go run test_faceit.go -key "YOUR_API_KEY" -url "https://www.faceit.com/en/cs2/room/1-..."
```

## Example Output

```
====================================
🎯 FACEIT API Test Tool
====================================

📂 Loading API key from config/secrets.json...
✅ API key loaded from secrets.json

🔗 Match URL: https://www.faceit.com/en/cs2/room/1-...

🚀 Initializing FACEIT client...
✅ Client initialized

📡 Fetching match data from FACEIT API...
------------------------------------
✅ Match data fetched successfully!

====================================
📊 Match Information
====================================

🆔 Match ID:      1-abc123-def456-...
📌 Status:        ONGOING
🏆 Competition:   FPL Europe
🎮 Game:          cs2

====================================
🔵 Team 1
====================================
Name:    Team A
Team ID: faction1
Score:   12
Logo:    https://assets.faceit.com/avatar/...

====================================
🔴 Team 2
====================================
Name:    Team B
Team ID: faction2
Score:   9
Logo:    https://assets.faceit.com/avatar/...

====================================
📺 GOTV Connection
====================================
Link:     connect 185.25.182.103:27015
Command:  connect 185.25.182.103:27015

====================================
✅ Test completed successfully!
====================================
```

## Finding Test Match URLs

To test the integration:

1. Go to [FACEIT](https://www.faceit.com/)
2. Find any ongoing CS2 match
3. Click on the match to open the room
4. Copy the URL from your browser
5. Use that URL in the test command

Example match URLs:
- FPL matches: `https://www.faceit.com/en/cs2/room/1-...`
- Tournament matches: `https://www.faceit.com/en/championship/.../1-...`

## Troubleshooting

**Error: "No API key found!"**
- Create `config/secrets.json` with your API key
- Or use `-key` flag to provide it directly

**Error: "could not extract match ID from URL"**
- Make sure URL contains `/room/1-...`
- Copy the full URL from browser

**Error: "FACEIT API error (status 401)"**
- API key is invalid or expired
- Get a new key from [FACEIT Developer Portal](https://developers.faceit.com/)

**Error: "FACEIT API error (status 404)"**
- Match ID doesn't exist
- Match might be too old (deleted from API)
- Verify URL is correct
