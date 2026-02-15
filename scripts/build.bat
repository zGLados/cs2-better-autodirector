@echo off
REM Wrapper for build.ps1
REM This allows double-clicking the .bat file to run the PowerShell script
echo Starting PowerShell build script...
powershell -ExecutionPolicy Bypass -File "%~dp0build.ps1"
exit /b %errorlevel%
