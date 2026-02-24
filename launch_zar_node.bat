@echo off
TITLE ZAR Blockchain Node
echo Starting ZAR Blockchain Node...

:: SET YOUR DUCKDNS CREDENTIALS HERE
set DUCKDNS_DOMAIN=zar-chain
set DUCKDNS_TOKEN=63d56e91-f094-4664-989d-df5a571b72d3

:: RUN THE NODE (Using full Go path for reliability)
"C:\Program Files\Go\bin\go.exe" run cmd/zar-node/main.go

pause
