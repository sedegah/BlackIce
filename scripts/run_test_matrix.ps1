$ErrorActionPreference = 'Stop'

$RootDir = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$LogDir = Join-Path $RootDir 'tests/logs'
if (-not (Test-Path $LogDir)) {
  New-Item -ItemType Directory -Path $LogDir | Out-Null
}

Write-Host "[1/4] go unit+integration tests"
Push-Location $RootDir
try {
  go test ./...
} finally {
  Pop-Location
}

Write-Host "[2/4] python unit tests"
python3 -m unittest discover -s (Join-Path $RootDir 'ml/inference') -p 'test_*.py'

Write-Host "[3/4] sidecar + runtime smoke"
$Sock = Join-Path $LogDir 'blackice-matrix.sock'
if (Test-Path $Sock) { Remove-Item $Sock -Force }
$PyLog = Join-Path $LogDir 'python_smoke.log'
$GoLog = Join-Path $LogDir 'go_smoke.log'

$py = Start-Process -FilePath python3 -ArgumentList @("$RootDir/ml/inference/server.py", '--socket', $Sock) -RedirectStandardOutput $PyLog -RedirectStandardError $PyLog -PassThru
Start-Sleep -Seconds 1

$go = Start-Process -FilePath go -ArgumentList @('run', './cmd/blackice', '--socket', $Sock, '--pps', '120000', '--window', '1s') -WorkingDirectory $RootDir -RedirectStandardOutput $GoLog -RedirectStandardError $GoLog -PassThru
Start-Sleep -Seconds 3

if (-not $go.HasExited) { Stop-Process -Id $go.Id -Force }
if (-not $py.HasExited) { Stop-Process -Id $py.Id -Force }

Write-Host "[4/4] quick assertion of mitigation output"
if (-not (Select-String -Path $GoLog -Pattern 'mitigation=' -Quiet)) {
  throw 'expected mitigation output missing'
}

Write-Host 'matrix complete'
