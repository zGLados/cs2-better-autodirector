@echo off
echo ============================================
echo Better Auto Observer - Quick Run
echo ============================================
echo.
echo Starting without building exe...
echo Press Ctrl+C to stop
echo.
set PATH=%PATH%;C:\TDM-GCC-64\bin
cd /d "%~dp0\.."
go run .
pause
