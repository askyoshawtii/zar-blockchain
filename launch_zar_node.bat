@echo off
title ZAR Blockchain Node - Automated Launcher
color 0A

echo ============================================
echo    ZAR BLOCKCHAIN - AUTOMATED NODE
echo ============================================
echo.

:: ─── Environment Variables ───
set DUCKDNS_DOMAIN=zar-chain
set DUCKDNS_TOKEN=63d56e91-f094-4664-989d-df5a571b72d3
set GO_PATH="C:\Program Files\Go\bin\go.exe"
set NODE_DIR=C:\Users\askyo\Desktop\zar-blockchain

:: ─── Navigate to project ───
cd /d %NODE_DIR%

:: ─── Start GitHub Auto-Sync in background ───
echo [AUTO-SYNC] Starting GitHub Auto-Sync daemon...
start /B powershell -WindowStyle Hidden -ExecutionPolicy Bypass -File "%NODE_DIR%\auto_sync.ps1"

:: ─── Start the Node ───
echo [NODE] Starting ZAR Blockchain Node with SSL...
echo [NODE] DuckDNS Domain: %DUCKDNS_DOMAIN%.duckdns.org
echo [NODE] Press Ctrl+C to stop.
echo.

%GO_PATH% run cmd/zar-node/main.go

pause
