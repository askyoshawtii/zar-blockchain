# ============================================
#  ZAR Blockchain - GitHub Auto-Sync Daemon
#  Runs in the background and pushes changes
#  to GitHub every 5 minutes automatically.
# ============================================

$ProjectDir = "C:\Users\askyo\Desktop\zar-blockchain"
$SyncInterval = 300  # Seconds (5 minutes)
$LogFile = "$ProjectDir\sync_log.txt"

function Write-Log {
    param([string]$Message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $entry = "[$timestamp] $Message"
    Add-Content -Path $LogFile -Value $entry
    Write-Host $entry
}

Set-Location $ProjectDir
Write-Log "Auto-Sync Daemon Started"

while ($true) {
    try {
        # Check if there are any changes
        $status = git status --porcelain 2>&1
        $hasChanges = ($status -ne $null -and $status.Length -gt 0)

        # Check if there are commits ahead of remote
        git fetch origin master 2>&1 | Out-Null
        $ahead = git rev-list --count origin/master..HEAD 2>&1

        if ($hasChanges) {
            Write-Log "Changes detected. Syncing..."

            # Stage all changes
            git add -A 2>&1 | Out-Null

            # Create auto-commit with timestamp
            $commitMsg = "auto-sync: $(Get-Date -Format 'yyyy-MM-dd HH:mm') | chaindata update"
            git commit -m $commitMsg 2>&1 | Out-Null

            # Pull remote changes first (rebase to keep history clean)
            git pull --rebase origin master 2>&1 | Out-Null

            # Push to GitHub
            $pushResult = git push origin master 2>&1
            Write-Log "Push complete: $pushResult"
        }
        elseif ($ahead -gt 0) {
            Write-Log "Unpushed commits found ($ahead ahead). Pushing..."
            git pull --rebase origin master 2>&1 | Out-Null
            git push origin master 2>&1 | Out-Null
            Write-Log "Push complete."
        }
        else {
            Write-Log "No changes. Sleeping $SyncInterval seconds..."
        }
    }
    catch {
        Write-Log "Sync Error: $_"
    }

    Start-Sleep -Seconds $SyncInterval
}
