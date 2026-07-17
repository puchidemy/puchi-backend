#!/bin/bash
set -euo pipefail 2>/dev/null || set -euo

AUTH_SECTION=$'\nauth:\n  auth_service_url: "http://auth-service.puchi-backend.svc.cluster.local:8080"\n  public_paths:\n    - /v1/health\n    - /v1/healthz'

for svc in user content game grading media notification; do
  echo "=== $svc ==="
  
  OLD=$(kubectl get configmap ${svc}-config -n puchi-backend -o jsonpath="{.data.config\.yaml}")
  
  if echo "$OLD" | grep -q "auth:"; then
    echo "Already has auth section, skipping"
    continue
  fi
  
  NEW="${OLD}${AUTH_SECTION}"
  
  ESCAPED=$(python3 -c "
import json, sys
data = sys.stdin.read()
print(json.dumps(data))
" <<< "$NEW")
  
  kubectl patch configmap ${svc}-config -n puchi-backend \
    -p "{\"data\":{\"config.yaml\":$ESCAPED}}" 2>&1
  
  echo ""
done
