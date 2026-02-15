@echo off
:: Build CS2 Better Auto Director Installer

echo ====================================================
echo   CS2 Better Auto Director - Installer Builder
echo ====================================================
echo.

powershell.exe -ExecutionPolicy Bypass -File "%~dp0build-installer.ps1"

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo Build failed!
    pause
    exit /b 1
)

pause
