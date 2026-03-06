# 🔐 Secrets Configuration

This file contains sensitive API keys and credentials that should **never** be committed to version control.

## Setup

1. Copy `secrets.example.json` to `secrets.json`:
   ```bash
   cp config/secrets.example.json config/secrets.json
   ```

2. Edit `config/secrets.json` with your actual API keys:
   ```json
   {
     "faceit_api_key": "your-actual-faceit-api-key-here"
   }
   ```

## Getting a FACEIT API Key

1. Go to [FACEIT Developer Portal](https://developers.faceit.com/)
2. Log in with your FACEIT account
3. Create a new application
4. Copy your API key
5. Paste it into `config/secrets.json`

## Security Notes

- ✅ `secrets.json` is automatically ignored by Git (see `.gitignore`)
- ✅ Never commit your actual API keys to the repository
- ✅ Share `secrets.example.json` as a template for other users
- ❌ Never share your `secrets.json` file publicly

## File Structure

```
config/
├── secrets.json              # Your actual secrets (NEVER commit!)
├── secrets.example.json      # Template file (safe to commit)
├── settings.json             # Application settings (can be committed)
└── ...
```

## Troubleshooting

**Error: "FACEIT client not initialized"**
- Make sure `config/secrets.json` exists
- Verify your API key is correctly set
- Restart the application

**No secrets file found**
- The application will log this message on startup
- FACEIT integration will not work until secrets.json is created
