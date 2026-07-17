#!/usr/bin/env bash
# Mở mỗi service một terminal tab/window và chạy kratos run.
# Ưu tiên: gnome-terminal / kitty / tmux. Fallback: background + logs.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
CONF_DIR="$ROOT/.dev/conf"
mkdir -p "$CONF_DIR"

if [[ -z "${LIMEN_SECRET:-}" || ${#LIMEN_SECRET} -ne 32 ]]; then
  export LIMEN_SECRET="dev-limen-secret-key-32bytes!!!!"
  echo "[dev] LIMEN_SECRET = local default (32 bytes)"
fi

if ! command -v kratos >/dev/null 2>&1; then
  echo "[dev] kratos CLI not found. Install: go install github.com/go-kratos/kratos/cmd/kratos/v2@latest"
  exit 1
fi

if [[ $# -gt 0 ]]; then
  SERVICES=("$@")
else
  SERVICES=(auth core learn media notification)
fi

http_port() {
  case "$1" in
    auth) echo 8080 ;;
    core) echo 8001 ;;
    learn) echo 8002 ;;
    media) echo 8003 ;;
    notification) echo 8004 ;;
  esac
}
grpc_port() {
  case "$1" in
    auth) echo "" ;;
    core) echo 9001 ;;
    learn) echo 9002 ;;
    media) echo 9003 ;;
    notification) echo 9004 ;;
  esac
}

write_conf() {
  local svc="$1" src="$ROOT/app/$svc/configs/config.yaml" dst="$CONF_DIR/$svc.yaml"
  local http grpc
  http="$(http_port "$svc")"
  grpc="$(grpc_port "$svc")"
  cp "$src" "$dst"
  if [[ "$svc" == "auth" ]]; then
    sed -i.bak -E "s|addr: \"[^\"]+\"|addr: \"0.0.0.0:${http}\"|" "$dst"
    # Public URL via Next rewrite (:3000) — cookie + OAuth callback same-origin with FE
    sed -i.bak -E 's|base_url: "[^"]+"|base_url: "http://localhost:3000"|' "$dst"
  else
    sed -i.bak -E "s|addr: 0\.0\.0\.0:8000|addr: 0.0.0.0:${http}|" "$dst"
    [[ -n "$grpc" ]] && sed -i.bak -E "s|addr: 0\.0\.0\.0:9000|addr: 0.0.0.0:${grpc}|" "$dst"
    sed -i.bak -E 's|auth_service_url: "[^"]+"|auth_service_url: "http://localhost:8080"|' "$dst"
  fi
  rm -f "$dst.bak"
}

OAUTH_ENV="$ROOT/.dev/oauth.env"
if [[ ! -f "$OAUTH_ENV" ]]; then
  echo "[dev] Thieu .dev/oauth.env — OAuth se thieu client_id."
  echo "      Copy scripts/dev/oauth.env.example -> .dev/oauth.env roi dien credentials."
fi

run_cmd() {
  local svc="$1"
  write_conf "$svc"
  local http conf_abs
  http="$(http_port "$svc")"
  # Absolute path: kratos run đổi cwd vào cmd/<svc>, relative ../../.dev sẽ sai
  conf_abs="$ROOT/.dev/conf/${svc}.yaml"
  cat <<EOF
export LIMEN_SECRET='$LIMEN_SECRET'
cd '$ROOT/app/$svc'
EOF
  if [[ "$svc" == "auth" && -f "$OAUTH_ENV" ]]; then
    # shellcheck disable=SC1090
    echo "set -a; source '$OAUTH_ENV'; set +a"
  fi
  cat <<EOF
echo "[puchi] $svc  http://localhost:$http"
# kratos run tự tìm ./cmd/<svc> — không truyền path (tránh nhân đôi)
kratos run -- -conf '$conf_abs'
EOF
}

open_tab() {
  local svc="$1"
  local script
  script="$(run_cmd "$svc")"

  if command -v gnome-terminal >/dev/null 2>&1; then
    gnome-terminal --title="puchi-$svc" --working-directory="$ROOT/app/$svc" -- bash -lc "$script; exec bash"
  elif command -v kitty >/dev/null 2>&1; then
    kitty --title "puchi-$svc" --directory "$ROOT/app/$svc" bash -lc "$script; exec bash" &
  elif command -v tmux >/dev/null 2>&1; then
    if ! tmux has-session -t puchi 2>/dev/null; then
      tmux new-session -d -s puchi -n "$svc" "bash -lc $(printf '%q' "$script")"
    else
      tmux new-window -t puchi -n "$svc" "bash -lc $(printf '%q' "$script")"
    fi
    echo "[dev] tmux window: puchi:$svc  (attach: tmux attach -t puchi)"
  else
    echo "[dev] No gnome-terminal/kitty/tmux — falling back to background logs"
    mkdir -p "$ROOT/.dev/logs" "$ROOT/.dev/pids"
    (
      cd "$ROOT/app/$svc"
      # shellcheck disable=SC2086
      nohup bash -lc "$script" >"$ROOT/.dev/logs/$svc.out.log" 2>"$ROOT/.dev/logs/$svc.err.log" &
      echo $! >"$ROOT/.dev/pids/$svc.pid"
    )
    echo "[dev] $svc pid=$(cat "$ROOT/.dev/pids/$svc.pid")"
  fi
}

for svc in "${SERVICES[@]}"; do
  echo "[dev] tab/window: puchi-$svc"
  open_tab "$svc"
done

echo ""
echo "auth :8080 | core :8001 | learn :8002 | media :8003 | notification :8004"
echo "Stop: Ctrl+C trong từng tab"
