#!/usr/bin/env bash
# Tabs: Ctrl+C. Optional: kill by port.
set -euo pipefail

if [[ "${1:-}" != "--kill-ports" ]]; then
  echo "Với run-all.sh (terminal tabs): Ctrl+C từng tab."
  echo "Force: ./scripts/dev/stop-all.sh --kill-ports"
  exit 0
fi

for p in 8080 8001 8002 8003 8004 9001 9002 9003 9004; do
  if command -v fuser >/dev/null 2>&1; then
    fuser -k "${p}/tcp" 2>/dev/null || true
  elif command -v lsof >/dev/null 2>&1; then
    pids=$(lsof -ti tcp:"$p" -sTCP:LISTEN 2>/dev/null || true)
    [[ -n "$pids" ]] && kill -9 $pids 2>/dev/null || true
  fi
  echo "[dev] cleared port $p"
done
echo "[dev] done"
