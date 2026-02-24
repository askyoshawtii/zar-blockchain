@echo off
title ZAR Blockchain Node - Automated Launcher
color 0A

echo ============================================
echo    ZAR BLOCKCHAIN - AUTOMATED NODE
echo ============================================
echo.

:: ─── Environment Variables ───
set "DUCKDNS_DOMAIN=zar-chain"
set "DUCKDNS_TOKEN=63d56e91-f094-4664-989d-df5a571b72d3"
set "NODE_DIR=C:\Users\askyo\Desktop\zar-blockchain"

:: ─── Navigate to project ───
cd /d "%NODE_DIR%"
if errorlevel 1 (
    echo [ERROR] Could not find project directory: %NODE_DIR%
    pause
    exit /b 1
)

:: ─── Verify Go is installed ───
where go >nul 2>nul
if errorlevel 1 (
    echo [ERROR] Go is not in PATH. Trying default location...
    set "PATH=%PATH%;C:\Program Files\Go\bin"
)

:: ─── Start GitHub Auto-Sync in background ───
echo [AUTO-SYNC] Starting GitHub Auto-Sync daemon...
start "" /B powershell -WindowStyle Hidden -ExecutionPolicy Bypass -File "%NODE_DIR%\auto_sync.ps1"

:: ─── Start the Node ───
echo [NODE] Starting ZAR Blockchain Node with SSL...
echo [NODE] DuckDNS Domain: %DUCKDNS_DOMAIN%.duckdns.org
echo [NODE] Bridge: BTC, ETH, SOL, LTC, DOGE, MATIC, XMR, TRX, BNB, PEPE, CELO, XRP, ADA
echo [NODE] Press Ctrl+C to stop.
echo.

"C:\Program Files\Go\bin\go.exe" run cmd/zar-node/main.go

echo.
echo ============================================
echo  NODE STOPPED. Check errors above.
echo ============================================
pause
