# 🎯 FACEIT Integration

The FACEIT integration allows you to automatically fetch match data from FACEIT and prepare for streaming competitive CS2 matches.

---

## 🌟 Features

- ✅ **Automatic Match Data Fetching** - Get team names and logos from FACEIT API
- ✅ **GOTV Server Information** - Retrieve GOTV connection details (when available)
- ✅ **Team Information** - Display team names, logos, and scores
- ✅ **Match Status** - See if match is live, ready, or finished
- ✅ **Competition Details** - View tournament/league information

---

## 🔧 Setup

### 1. Get FACEIT API Key

1. Visit [FACEIT Developer Portal](https://developers.faceit.com/)
2. Log in with your FACEIT account
3. Click **"Create App"** or **"My Apps"**
4. Create a new application:
   - **Name**: `CS2 Auto Director` (or any name)
   - **Description**: Personal use for match streaming
   - **Redirect URLs**: Leave empty (not needed)
5. Copy your **API Key**

### 2. Configure API Key

Create or edit `config/secrets.json`:

```json
{
  "faceit_api_key": "your-actual-api-key-here"
}
```

**Security Note:** This file is automatically excluded from Git (see `.gitignore`). Never commit your API key!

For more details, see [SECRETS.md](SECRETS.md).

### 3. Launch Application

The FACEIT client will be automatically initialized on startup if the API key is configured.

```bash
# Start the GUI
./cs2-better-autodirector.exe

# Or development mode
cd gui
wails dev
```

---

## 🧪 Testing the Integration

Before using FACEIT integration in the full application, you can test it standalone:

### Quick Test Tool

Use the test script to verify your API key and fetch match data:

```bash
cd scripts
go run test_faceit.go -url "https://www.faceit.com/en/cs2/room/1-YOUR-MATCH-ID"
```

**What it does:**
- ✅ Loads API key from `config/secrets.json`
- ✅ Fetches match data from FACEIT API
- ✅ Displays all match information (teams, scores, GOTV link)
- ✅ Shows full JSON response for debugging

**Example Output:**
```
====================================
🎯 FACEIT API Test Tool
====================================

📂 Loading API key from config/secrets.json...
✅ API key loaded from secrets.json

🔗 Match URL: https://www.faceit.com/en/cs2/room/...

📡 Fetching match data from FACEIT API...
✅ Match data fetched successfully!

====================================
📊 Match Information
====================================
🆔 Match ID:      1-abc123...
📌 Status:        ONGOING
🏆 Competition:   FPL Europe
...
```

**Full test documentation:** [Testing Guide](FACEIT_TESTING.md)

---

## 🎮 Usage

### Fetching Match Data

1. **Find Match Room URL**
   - Go to FACEIT and find the match you want to stream
   - Copy the match room URL, e.g.:
     ```
     https://www.faceit.com/en/cs2/room/1-a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
     ```

2. **In the Auto Director Dashboard**
   - Locate the **"FACEIT Match Integration"** widget (orange header)
   - Paste the match room URL into the input field
   - Click **"Fetch Match Data"**

3. **View Match Information**
   - Team 1 and Team 2 names and logos appear
   - Current score is displayed
   - Match status (Live, Ready, Finished, etc.)
   - Competition name
   - GOTV connection link (if available)

### Connecting to GOTV

1. **Copy GOTV Link**
   - Click on the GOTV link field to select and copy it
   - The link is formatted as: `connect IP:PORT`

2. **Connect in CS2**
   - Open CS2 console (`~` key)
   - Paste the connect command:
     ```
     connect 185.25.182.103:27015
     ```
   - Press Enter to connect

3. **Start Auto Director**
   - Once connected as spectator in GOTV
   - Execute `exec spectator_bindings` in console
   - Click **"Start Auto Director"** button
   - The system will automatically control the camera

---

## 📊 Match Data Structure

### What Data is Fetched

```json
{
  "match_id": "1-a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6",
  "gotv_link": "connect 185.25.182.103:27015",
  "status": "ONGOING",
  "competition": "FPL Europe",
  "team1": {
    "name": "Team A",
    "logo": "https://assets.faceit.com/avatar/...",
    "score": 12
  },
  "team2": {
    "name": "Team B",
    "logo": "https://assets.faceit.com/avatar/...",
    "score": 9
  }
}
```

### Match Status Types

| Status | Meaning |
|--------|---------|
| `READY` | Match configured, waiting to start |
| `ONGOING` | Match is currently live |
| `FINISHED` | Match has ended |
| `CANCELLED` | Match was cancelled |
| `CONFIGURING` | Match is being set up |

---

## 🔌 API Integration Details

### Endpoint Used

```
GET https://open.faceit.com/data/v4/matches/{match_id}
```

### Authentication

Uses Bearer token authentication:
```
Authorization: Bearer YOUR_API_KEY
```

### Rate Limits

FACEIT API has rate limits:
- **Free tier**: ~300 requests per hour
- For personal use, this is more than sufficient

### Data Privacy

- Only public match data is accessed
- No player personal information is stored
- API key is stored locally only

---

## 🎬 Streaming Workflow

### Complete Streaming Setup

1. **Before Match Day**
   ```
   ✓ Configure FACEIT API key
   ✓ Test Auto Director with a public match
   ✓ Set up OBS/Streaming software
   ```

2. **Match Day Preparation**
   ```
   ✓ Fetch FACEIT match data
   ✓ Note GOTV connection details
   ✓ Start streaming software
   ```

3. **Go Live**
   ```
   ✓ Connect to GOTV server
   ✓ Execute spectator bindings
   ✓ Start Auto Director
   ✓ Begin stream
   ```

4. **During Match**
   ```
   ✓ Auto Director handles camera
   ✓ Monitor dashboard for stats
   ✓ Check encounter priorities in logs
   ```

---

## ⚠️ Limitations & Known Issues

### GOTV Link Availability

- **Live Matches**: GOTV link might not be available immediately
  - FACEIT API doesn't always expose GOTV info directly
  - May require match to start before server info is available
  - Alternative: Check FACEIT match page manually

- **Finished Matches**: 
  - Demo URLs are typically provided
  - GOTV link shows replay file instead of live server

### Workarounds

If GOTV link is not available:
1. Check the FACEIT match room page manually
2. Look for server info in match chat
3. Wait for match to start - link may appear after first round

---

## 🛠️ Troubleshooting

### "FACEIT client not initialized"

**Cause**: No API key configured or invalid key.

**Solution**:
```bash
# Create secrets file
cp config/secrets.example.json config/secrets.json

# Edit with your API key
nano config/secrets.json

# Restart application
```

### "Failed to fetch match data"

**Possible Causes**:
1. Invalid match room URL
2. Match ID extraction failed
3. API key is invalid
4. Network connectivity issues

**Solutions**:
- Verify URL format: `https://www.faceit.com/.../room/1-...`
- Check API key is correct
- Test API key: `curl -H "Authorization: Bearer YOUR_KEY" https://open.faceit.com/data/v4/auth/test`
- Check logs for detailed error messages

### "Could not extract match ID from URL"

**Cause**: URL format doesn't match expected pattern.

**Expected Format**:
```
https://www.faceit.com/en/cs2/room/1-abc123-def456-...
https://www.faceit.com/en/csgo/room/1-abc123-def456-...
```

**Solution**: Copy the full match room URL from the browser address bar.

### Team logos not loading

**Cause**: Image URL blocked or unavailable.

**Note**: This is cosmetic only and doesn't affect functionality.

---

## 🔮 Future Enhancements

Planned features for FACEIT integration:

- [ ] **Auto-refresh** - Periodically update match data while running
- [ ] **WebSocket Support** - Real-time updates during match
- [ ] **Player Stats** - Display individual player performance
- [ ] **Multiple Matches** - Queue and switch between multiple matches
- [ ] **HUD Export** - Export match data for OBS browser sources
- [ ] **Veto Information** - Show map picks and bans
- [ ] **Live Score Updates** - Real-time score tracking

---

## 📝 API Response Example

### Full Match Response

```json
{
  "match_id": "1-a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6",
  "status": "ONGOING",
  "started_at": 1709737200,
  "competition_name": "FPL Europe",
  "game": "cs2",
  "teams": {
    "faction1": {
      "name": "Team A",
      "avatar": "https://assets.faceit.com/avatar/...",
      "roster": [
        {
          "player_id": "...",
          "nickname": "player1"
        }
      ]
    },
    "faction2": {
      "name": "Team B",
      "avatar": "https://assets.faceit.com/avatar/...",
      "roster": [...]
    }
  },
  "results": {
    "score": {
      "faction1": 12,
      "faction2": 9
    }
  }
}
```

---

## 🔗 Related Documentation

- **[Testing Guide](FACEIT_TESTING.md)** - Test FACEIT integration standalone
- [Secrets Configuration](SECRETS.md) - API key setup
- [Main README](../README.md) - General usage
- [Camera Priority](CAMERA_PRIORITY.md) - Auto Director algorithm

---

## 🌐 External Resources

- [FACEIT Developer Portal](https://developers.faceit.com/)
- [FACEIT API Documentation](https://developers.faceit.com/docs/tools/data-api)
- [FACEIT Support](https://support.faceit.com/)

---

**Need Help?**
- Check the [troubleshooting section](#-troubleshooting) above
- Open an [issue on GitHub](https://github.com/zGLados/cs2-better-autodirector/issues)
- Read the [API documentation](https://developers.faceit.com/docs/)
