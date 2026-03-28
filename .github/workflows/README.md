﻿# GitHub Actions Workflows

Automated build pipelines for creating Windows installers on GitHub's servers.

## Workflow: `build-installer.yml`

**Platform:** Windows

**Triggers:**
- **Version tags** (`v*.*.*`): Builds installer + creates GitHub Release
- **Manual dispatch**: Actions → "Build Windows Installer" → Run workflow

**What it does:**
1. Sets up Go 1.22 + Node.js 20 + Wails CLI
2. Installs Inno Setup (via Chocolatey, ~30 sec)
3. Builds GUI application (`wails build -skipbindings`)
4. Compiles installer (`ISCC.exe installer.iss`)
5. Uploads artifacts (90 days retention):
   - Portable EXE: `cs2-better-autodirector.exe`
   - Installer: `CS2BetterAutoDirector-Setup.exe`
6. Creates GitHub Release (if triggered by tag)

**Build time:** ~5-7 minutes

**Output:** 
- `CS2BetterAutoDirector-Setup.exe` (Installer with auto-config)
- `cs2-better-autodirector.exe` (Portable, no installation needed)

---
## Usage

### Create Release

```bash
# 1. Update version in code/docs, commit changes
git add .
git commit -m "Release v0.3.1"
git push origin dev

# 2. Create and push tag
git tag -a v0.3.1 -m "Version 0.3.1"
git push origin v0.3.1

# → Installer automatically attached to GitHub Release
```

### Manual Build (Testing)

**Windows:**
1. Go to: Repository → Actions → "Build Windows Installer"
2. Click "Run workflow" → Select branch → Run
3. Wait ~5-7 minutes
4. Download from Artifacts section

---

## Workflow Behavior

| Trigger | Windows Build? | Creates Release? |
|---------|----------------|------------------|
| Push to `dev` | ❌ No | ❌ No |
| Pull Request | ❌ No | ❌ No |
| Tag `v*.*.*` | ✅ Yes | ✅ Yes |
| Manual Run | ✅ Yes | ❌ No |

**Note:** Builds only on tags to save GitHub Actions minutes. For local testing on Windows:
- Windows: `.\scripts\build.ps1` or `.\scripts\build-installer.ps1`

---

## Troubleshooting

### Windows Builds

**Build fails at Wails:**
- Ensure `gui/go.mod` and `gui/go.sum` are committed
- Check Go version matches workflow (1.22)

**Installer not created:**
- Verify `gui/build/bin/cs2-better-autodirector.exe` exists after build
- Check Inno Setup paths in `scripts/installer.iss`

**Release not created:**
- Ensure tag starts with `v` (e.g., `v0.2.3`, not `0.2.3`)
- Check repository has `contents: write` permission

### Linux Builds

**Build fails at dependencies:**
- Verify GTK3 and WebKit2GTK are available in Ubuntu repos
- Check if `apt-get update` succeeded
- Ubuntu 24.04+ requires `libwebkit2gtk-4.1-dev` (not 4.0)
- If using Ubuntu 24.04+, ensure the webkit2gtk-4.0.pc symlink is created

**Build fails at Wails:**
- Ensure `gui/go.mod` and `gui/go.sum` are committed
- Check Go version matches workflow (1.22)
- Verify pkg-config can find webkit2gtk-4.0 (via symlink on Ubuntu 24.04+)
- Check if symlink exists: `ls -la /usr/lib/x86_64-linux-gnu/pkgconfig/webkit2gtk-4.0.pc`

**Package not created:**
- Verify `gui/build/bin/cs2-better-autodirector` exists after build
- Check if `scripts/build-package.sh` is executable
- Review build logs for tar command errors

---

For detailed build instructions, silent installation, and troubleshooting, see **[docs/BUILD.md](../docs/BUILD.md)**.
