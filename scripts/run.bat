@echo off
echo ============================================
echo CS2 Better Auto Director - Quick Run (GUI)
echo ============================================
echo.
echo Starting in development mode (hot reload)...
echo Press Ctrl+C to stop
echo.
cd /d "%~dp0\..\cs2-autodirector-gui"
wails dev
pause
