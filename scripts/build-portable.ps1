# Builds the portable version: build\CS2BetterAutoDirector-Portable.zip
# Contents: cs2-better-autodirector.exe, config\*.cfg, COMO-USAR.txt, "Abrir em modo debug.bat"
# Usage (from the project folder):  .\scripts\build-portable.ps1

$ErrorActionPreference = "Stop"
$root = Split-Path $PSScriptRoot -Parent

Write-Host "[1/3] Compiling the app..." -ForegroundColor Green
Push-Location "$root\gui"
try {
    wails build -skipbindings
    if ($LASTEXITCODE -ne 0) { throw "wails build failed" }
} finally {
    Pop-Location
}

Write-Host "[2/3] Assembling the portable folder..." -ForegroundColor Green
$name = "CS2BetterAutoDirector-Portable"
$out = "$root\build\$name"
if (Test-Path $out) { Remove-Item $out -Recurse -Force }
New-Item -ItemType Directory -Path "$out\config" -Force | Out-Null
Copy-Item "$root\gui\build\bin\cs2-better-autodirector.exe" $out
Copy-Item "$root\config\gamestate_integration_autodirector.cfg" "$out\config"
Copy-Item "$root\config\spectator_bindings.cfg" "$out\config"
Copy-Item "$root\COMO-USAR.txt" $out
$bat = "@echo off`r`nstart """" ""%~dp0cs2-better-autodirector.exe"" -v`r`n"
[System.IO.File]::WriteAllText("$out\Abrir em modo debug.bat", $bat, [System.Text.Encoding]::ASCII)

Write-Host "[3/3] Creating the zip..." -ForegroundColor Green
$zip = "$root\build\$name.zip"
if (Test-Path $zip) { Remove-Item $zip -Force }
Compress-Archive -Path "$out\*" -DestinationPath $zip

Write-Host ""
Write-Host "Done: $zip" -ForegroundColor Cyan
