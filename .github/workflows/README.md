# GitHub Actions Workflows

This folder contains automated build workflows for GitHub Actions.

## Available Workflows

### `build-installer.yml` - Windows Installer Build

**Automatically builds the Windows installer on GitHub's servers.**

#### Trigger Methods:

1. **Automatic (Every Push to Dev):**
   ```bash
   git add .
   git commit -m "Update feature X"
   git push origin dev
   ```
   → Builds installer automatically, uploads as artifact (no release)
   → Perfect for testing changes

2. **Automatic (Git Tags - Official Releases):**
   ```bash
   git tag v3.1.0
   git push origin v3.1.0
   ```
   → Builds installer + creates GitHub Release with installer attached
   → Perfect for official releases

3. **Automatic (Pull Requests):**
   ```bash
   # Create PR on GitHub
   ```
   → Builds installer to verify PR doesn't break the build
   → No release created, only artifact

4. **Manual (GitHub UI):**
   - Go to: Repository → Actions → "Build Windows Installer"
   - Click "Run workflow" → Select branch → Run
   → Creates an artifact (downloadable for 90 days)
   → No release created

### Workflow Behavior:

| Trigger | Builds Installer? | Creates Artifact? | Creates Release? |
|---------|-------------------|-------------------|------------------|
| Push to `dev` | ✅ Yes | ✅ Yes (90 days) | ❌ No |
| Pull Request | ✅ Yes | ✅ Yes (90 days) | ❌ No |
| Tag `v*` | ✅ Yes | ✅ Yes (90 days) | ✅ Yes (permanent) |
| Manual Run | ✅ Yes | ✅ Yes (90 days) | ❌ No |

### Recommended Workflow:

**Daily development:**
```bash
# Make changes, commit, push to dev
git add .
git commit -m "Add new feature"
git push origin dev
# → Automatic build, download from Actions → Artifacts
```

**Official release:**
```bash
# 1. Update CHANGELOG.md, version numbers, etc.
git add .
git commit -m "Release v3.2.0"
git push origin dev

# 2. Create and push tag
git tag -a v3.2.0 -m "Version 3.2.0 - Description"
git push origin v3.2.0
# → Automatic build + Release on GitHub
```

#### What it does:

1. ✅ Sets up Go 1.22
2. ✅ Sets up Node.js 20
3. ✅ Installs Wails CLI
4. ✅ Installs Inno Setup 6 (via Chocolatey - fast!)
5. ✅ Builds GUI application (`wails build -skipbindings`)
6. ✅ Compiles installer (`ISCC.exe installer.iss`)
7. ✅ Uploads `CS2BetterAutoDirector-Setup.exe` as artifact
8. ✅ Creates GitHub Release (if triggered by tag)

#### Build Time:

- First run: ~5-7 minutes (installs all dependencies)
- Cached runs: ~3-5 minutes (with dependency caching)
- Inno Setup install: ~30-60 seconds (via Chocolatey)

#### Artifacts:

- **Name**: `CS2BetterAutoDirector-Setup`
- **File**: `CS2BetterAutoDirector-Setup.exe` (~5-7 MB)
- **Retention**: 90 days
- **Download**: Actions → Workflow run → Artifacts section

#### Release Assets:

When triggered by a version tag (e.g., `v3.1.0`), the installer is automatically:
- Attached to the GitHub Release
- Available for download at: `https://github.com/USER/REPO/releases/tag/v3.1.0`
- Permanent (no expiration)

## Usage Examples

### Create a new release:

```bash
# 1. Update version in code/docs
# 2. Commit changes
git add .
git commit -m "Release v3.1.0"

# 3. Create and push tag
git tag -a v3.1.0 -m "Version 3.1.0 - Professional Installer"
git push origin main
git push origin v3.1.0

# 4. Wait for GitHub Actions to complete (~8 min)
# 5. Check GitHub Releases page
```

### Manual build without release:

1. Go to: https://github.com/YOUR_USERNAME/cs2-better-autodirector/actions
2. Click "Build Windows Installer" workflow
3. Click "Run workflow" dropdown
4. Select branch (usually `main`)
5. Click green "Run workflow" button
6. Wait for completion (~8 min)
7. Download artifact from workflow run page

## Benefits

✅ **No local dependencies**: Don't need Inno Setup or Node.js installed locally
✅ **Consistent builds**: Same environment every time
✅ **Automatic releases**: Tag → Installer → Release (fully automated)
✅ **Build artifacts**: Download installers from any commit/branch
✅ **Build logs**: Full transparency, debug build issues easily
✅ **Free for public repos**: GitHub Actions is free for public repositories

## Limitations

⚠️ **Windows only**: Currently builds only for Windows (runs on `windows-latest`)
⚠️ **Build time**: Takes longer than local builds (cold cache)
⚠️ **Internet required**: Downloads dependencies every time (can be optimized with caching)

## Optimization Ideas

**Add dependency caching:**
```yaml
- name: Cache Go modules
  uses: actions/cache@v4
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

**Add Node.js caching:**
```yaml
- name: Cache Node modules
  uses: actions/cache@v4
  with:
    path: gui/frontend/node_modules
    key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
```

This would reduce build time from ~8 minutes to ~3-4 minutes on subsequent runs.

## Troubleshooting

**Workflow fails at Inno Setup installation:**
- Check if Inno Setup download URL is still valid
- Update download link in `build-installer.yml`

**Wails build fails:**
- Ensure `gui/go.mod` and `gui/go.sum` are committed
- Ensure `gui/frontend/package.json` is committed
- Check Go and Node.js versions in workflow

**Installer not created:**
- Check that `gui/build/bin/cs2-better-autodirector.exe` exists after Wails build
- Verify Inno Setup path: `C:\Program Files (x86)\Inno Setup 6\ISCC.exe`
- Check `scripts/installer.iss` paths are relative and correct

## Monitoring

View build status:
- **Badge**: Add to README.md: `![Build](https://github.com/USER/REPO/actions/workflows/build-installer.yml/badge.svg)`
- **Email**: Configure in GitHub Settings → Notifications
- **Slack/Discord**: Use GitHub Actions integrations

---

For more information on GitHub Actions, see: https://docs.github.com/en/actions
