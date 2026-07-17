#Requires -Version 5.1
<#
.SYNOPSIS
  Goi y dung stack local (tabs wt dung Ctrl+C).
  Them -KillPorts de kill process dang listen port 8080/8001-8004.
#>
param(
  [switch]$KillPorts
)

Write-Host "Với run-all.ps1 (Windows Terminal tabs):" -ForegroundColor White
Write-Host "  - Ctrl+C trong từng tab, hoặc đóng tab / cửa sổ wt"
Write-Host ""

if (-not $KillPorts) {
  Write-Host "Force kill theo port: .\scripts\dev\stop-all.ps1 -KillPorts" -ForegroundColor DarkGray
  exit 0
}

$ports = @(8080, 8001, 8002, 8003, 8004, 9001, 9002, 9003, 9004)
foreach ($p in $ports) {
  $conns = Get-NetTCPConnection -LocalPort $p -State Listen -ErrorAction SilentlyContinue
  foreach ($c in $conns) {
    $procId = $c.OwningProcess
    if ($procId -and $procId -ne 0) {
      $name = (Get-Process -Id $procId -ErrorAction SilentlyContinue).ProcessName
      Write-Host "[dev] kill pid=$procId ($name) port=$p" -ForegroundColor Yellow
      Stop-Process -Id $procId -Force -ErrorAction SilentlyContinue
    }
  }
}
Write-Host "[dev] done" -ForegroundColor Green
