#Requires -Version 5.1
<#
.SYNOPSIS
  Mo tabs trong Windows Terminal hien tai — moi tab chay kratos run.

.EXAMPLE
  .\scripts\dev\run-all.ps1
  pwsh -File .\scripts\dev\run-all.ps1
#>
param(
  [string[]]$Services = @("auth", "core", "learn", "media", "notification"),
  [string]$Window = "0"
)

$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "../..")).Path
$ConfDir = Join-Path $Root ".dev\conf"
$RunDir = Join-Path $Root ".dev\run"
New-Item -ItemType Directory -Force -Path $ConfDir, $RunDir | Out-Null

function Resolve-ShellPath {
  # 1) Shell dang chay script nay (powershell.exe hoac pwsh.exe)
  if ($PSHOME) {
    foreach ($name in @("pwsh.exe", "powershell.exe")) {
      $p = Join-Path $PSHOME $name
      if (Test-Path -LiteralPath $p) { return (Resolve-Path $p).Path }
    }
  }
  # 2) Process hien tai
  try {
    $procPath = (Get-Process -Id $PID).Path
    if ($procPath -and (Test-Path -LiteralPath $procPath)) { return $procPath }
  } catch {}

  # 3) Tim tren may
  foreach ($c in @(
      "$env:ProgramFiles\PowerShell\7\pwsh.exe"
      "$env:ProgramFiles\PowerShell\7-preview\pwsh.exe"
      "$env:WINDIR\System32\WindowsPowerShell\v1.0\powershell.exe"
    )) {
    if (Test-Path -LiteralPath $c) { return $c }
  }

  $cmd = Get-Command pwsh, powershell -ErrorAction SilentlyContinue |
    Where-Object { $_.Source -and $_.Source -notmatch '\\WindowsApps\\' } |
    Select-Object -First 1
  if ($cmd) { return $cmd.Source }

  return $null
}

if (-not (Get-Command wt.exe -ErrorAction SilentlyContinue)) {
  Write-Host "[dev] Khong tim thay wt.exe (Windows Terminal)." -ForegroundColor Red
  exit 1
}

$shell = Resolve-ShellPath
if (-not $shell) {
  Write-Host "[dev] Khong tim thay powershell.exe / pwsh.exe" -ForegroundColor Red
  exit 1
}
Write-Host "[dev] shell = $shell" -ForegroundColor DarkGray

if (-not (Get-Command kratos -ErrorAction SilentlyContinue)) {
  Write-Host "[dev] Khong tim thay kratos. go install github.com/go-kratos/kratos/cmd/kratos/v2@latest" -ForegroundColor Red
  exit 1
}

$limen = $env:LIMEN_SECRET
if (-not $limen -or $limen.Length -ne 32) {
  $limen = "dev-limen-secret-key-32bytes!!!!"
  Write-Host "[dev] LIMEN_SECRET = local default (32 bytes)" -ForegroundColor Yellow
}

function Read-DotEnv([string]$path) {
  $map = @{}
  if (-not (Test-Path -LiteralPath $path)) { return $map }
  foreach ($line in Get-Content -LiteralPath $path) {
    $t = $line.Trim()
    if (-not $t -or $t.StartsWith("#")) { continue }
    $i = $t.IndexOf("=")
    if ($i -lt 1) { continue }
    $k = $t.Substring(0, $i).Trim()
    $v = $t.Substring($i + 1).Trim()
    if (($v.StartsWith('"') -and $v.EndsWith('"')) -or ($v.StartsWith("'") -and $v.EndsWith("'"))) {
      $v = $v.Substring(1, $v.Length - 2)
    }
    $map[$k] = $v
  }
  return $map
}

$oauthEnvPath = Join-Path $Root ".dev\oauth.env"
$oauth = Read-DotEnv $oauthEnvPath
if ($oauth.Count -eq 0) {
  Write-Host "[dev] Thieu .dev/oauth.env - OAuth se thieu client_id." -ForegroundColor Yellow
  Write-Host "      Copy scripts/dev/oauth.env.example -> .dev/oauth.env roi dien credentials." -ForegroundColor DarkGray
} else {
  $oauthKeyCount = $oauth.Count
  Write-Host "[dev] loaded OAuth env from .dev/oauth.env - $oauthKeyCount keys" -ForegroundColor DarkGray
}

$ports = @{
  auth         = @{ http = 8080; grpc = $null }
  core         = @{ http = 8001; grpc = 9001 }
  learn        = @{ http = 8002; grpc = 9002 }
  media        = @{ http = 8003; grpc = 9003 }
  notification = @{ http = 8004; grpc = 9004 }
}

function Write-LocalConfig([string]$svc) {
  $src = Join-Path $Root "app\$svc\configs\config.yaml"
  if (-not (Test-Path $src)) { throw "Missing config: $src" }
  $text = [System.IO.File]::ReadAllText($src)
  $http = $ports[$svc].http
  $grpc = $ports[$svc].grpc
  if ($svc -eq "auth") {
    $text = $text -replace 'addr:\s*"[^"]+"', ("addr: `"0.0.0.0:{0}`"" -f $http)
    # Public URL via Next rewrite (:3000) — cookie + OAuth callback same-origin with FE
    $text = $text -replace 'base_url:\s*"[^"]+"', 'base_url: "http://localhost:3000"'
  } else {
    $text = $text -replace 'addr:\s*0\.0\.0\.0:8000', ("addr: 0.0.0.0:{0}" -f $http)
    if ($null -ne $grpc) {
      $text = $text -replace 'addr:\s*0\.0\.0\.0:9000', ("addr: 0.0.0.0:{0}" -f $grpc)
    }
    $text = $text -replace 'auth_service_url:\s*"[^"]+"', 'auth_service_url: "http://localhost:8080"'
    # Local introspect + DB can exceed default 1s under cold start
    $text = $text -replace 'timeout:\s*1s', 'timeout: 5s'
  }
  [System.IO.File]::WriteAllText((Join-Path $ConfDir "$svc.yaml"), $text)
}

function Escape-PsSingle([string]$s) {
  if ($null -eq $s) { return "" }
  return $s.Replace("'", "''")
}

function Write-Launcher([string]$svc) {
  $http = $ports[$svc].http
  # Absolute path: kratos run doi cwd vao cmd/<svc>, relative ../../.dev se sai
  $confAbs = Join-Path $ConfDir "$svc.yaml"
  $launcher = Join-Path $RunDir "$svc.ps1"
  $work = Join-Path $Root "app\$svc"

  $envBlock = "`$env:LIMEN_SECRET = '$(Escape-PsSingle $limen)'`n"
  if ($svc -eq "auth") {
    foreach ($key in @(
        "GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
        "FACEBOOK_CLIENT_ID", "FACEBOOK_CLIENT_SECRET",
        "TIKTOK_CLIENT_KEY", "TIKTOK_CLIENT_SECRET"
      )) {
      if ($oauth.ContainsKey($key) -and $oauth[$key]) {
        $envBlock += "`$env:$key = '$(Escape-PsSingle $oauth[$key])'`n"
      }
    }
  }

  $body = @"
`$ErrorActionPreference = 'Continue'
$envBlock
Set-Location -LiteralPath '$work'
Write-Host "[puchi] $svc  http://localhost:$http" -ForegroundColor Cyan
Write-Host "[puchi] kratos run -- -conf $confAbs" -ForegroundColor DarkGray
# kratos run tu app/<svc>: tu tim ./cmd/<svc> — KHONG truyen path (se nhan doi cmd/<svc>/cmd/<svc>)
kratos run -- -conf "$confAbs"
"@
  $utf8Bom = New-Object System.Text.UTF8Encoding $true
  [System.IO.File]::WriteAllText($launcher, $body, $utf8Bom)
  return $launcher
}

$wtArgs = @("-w", $Window)
$first = $true

foreach ($svc in $Services) {
  if (-not $ports.ContainsKey($svc)) {
    Write-Host "[dev] unknown service: $svc (skip)" -ForegroundColor Red
    continue
  }
  Write-LocalConfig $svc
  $launcher = Write-Launcher $svc
  $workDir = Join-Path $Root "app\$svc"

  if (-not $first) { $wtArgs += ";" }
  $first = $false

  $wtArgs += @(
    "new-tab",
    "--title", "puchi-$svc",
    "-d", $workDir,
    "--",
    $shell,
    "-NoExit",
    "-NoProfile",
    "-ExecutionPolicy", "Bypass",
    "-File", $launcher
  )
  Write-Host "[dev] tab puchi-$svc  :$($ports[$svc].http)" -ForegroundColor Green
}

if ($first) {
  Write-Host "[dev] no services" -ForegroundColor Red
  exit 1
}

Write-Host ""
Write-Host "Mo tabs trong wt (-w $Window)..." -ForegroundColor White
Write-Host "Dung: Ctrl+C tung tab" -ForegroundColor DarkGray
Write-Host ""

& wt.exe @wtArgs
