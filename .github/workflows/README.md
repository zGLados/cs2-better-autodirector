# GitHub Actions Workflows

Automated build pipeline for creating Windows installers on GitHub's servers.

## Workflow: `build-installer.yml`

**Triggers:**
- **Version tags** (`v*.*.*`): Builds installer + creates GitHub Release
- **Manual dispatch**: Actions → "Build Windows Installer" → Run workflow

**What it does:**
1. Sets up Go 1.22 + Node.js 20 + Wails CLI
2. Installs Inno Setup (via Chocolatey, ~30 sec)
3. Builds GUI application (`wails build -skipbindings`)
4. Compiles installer (`ISCC.exe installer.iss`)
5. Uploads artifact (90 days retention)
6. Creates GitHub Release (if triggered by tag)

**Build time:** ~5-7 minutes

---

## Usage

### Create Release

```bash
# 1. Update version in code/docs, commit changes
git add .
git commit -m "Release v3.2.0"
git push origin dev

# 2. Create and push tag
git tag -a v3.2.0 -m "Version 3.2.0 - Description"
git push origin v3.2.0

# → Installer automatically attached to GitHub Release
```

### Manual Build (Testing)

1. Go to: Repository → Actions → "Build Windows Installer"
2. Click "Run workflow" → Select branch → Run
3. Wait ~5-7 minutes
4. Download from Artifacts section

---

## Workflow Behavior

| Trigger | Builds? | Creates Release? |
|---------|---------|------------------|
| Push to `dev` | ❌ No | ❌ No |
| Pull Request | ❌ No | ❌ No |
| Tag `v*.*.*` | ✅ Yes | ✅ Yes |
| Manual Run | ✅ Yes | ❌ No |

**Note:** Builds only on tags to save GitHub Actions minutes. For local testing, use `.\scripts\build-installer.bat`.

---

## Troubleshooting

**Build fails at Wails:**
- Ensure `gui/go.mod` and `gui/go.sum` are committed
- Check Go version matches workflow (1.22)

**Installer not created:**
- Verify `gui/build/bin/cs2-better-autodirector.exe` exists after build
- Check Inno Setup paths in `scripts/installer.iss`

**Release not created:**
- Ensure tag starts with `v` (e.g., `v0.1.0`, not `0.1.0`)
- Check repository has `contents: write` permission

---

For detailed build instructions, silent installation, and troubleshooting, see **[docs/BUILD.md](../docs/BUILD.md)**.
