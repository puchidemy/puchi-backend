#!/bin/bash
set -e

G_ID=$(kubectl get secret oauth-credentials -n puchi-frontend -o jsonpath='{.data.GOOGLE_CLIENT_ID}' | base64 -d)
G_SEC=$(kubectl get secret oauth-credentials -n puchi-frontend -o jsonpath='{.data.GOOGLE_CLIENT_SECRET}' | base64 -d)
F_ID=$(kubectl get secret oauth-credentials -n puchi-frontend -o jsonpath='{.data.FACEBOOK_CLIENT_ID}' | base64 -d)
F_SEC=$(kubectl get secret oauth-credentials -n puchi-frontend -o jsonpath='{.data.FACEBOOK_CLIENT_SECRET}' | base64 -d)
T_KEY=$(kubectl get secret oauth-credentials -n puchi-frontend -o jsonpath='{.data.TIKTOK_CLIENT_KEY}' | base64 -d)
T_SEC=$(kubectl get secret oauth-credentials -n puchi-frontend -o jsonpath='{.data.TIKTOK_CLIENT_SECRET}' | base64 -d)

echo "lens: G=${#G_ID} F=${#F_ID} T=${#T_KEY}"

kubectl create secret generic auth-oauth-credentials -n puchi-backend \
  --from-literal=GOOGLE_CLIENT_ID="$G_ID" \
  --from-literal=GOOGLE_CLIENT_SECRET="$G_SEC" \
  --from-literal=FACEBOOK_CLIENT_ID="$F_ID" \
  --from-literal=FACEBOOK_CLIENT_SECRET="$F_SEC" \
  --from-literal=TIKTOK_CLIENT_KEY="$T_KEY" \
  --from-literal=TIKTOK_CLIENT_SECRET="$T_SEC" \
  --dry-run=client -o yaml | kubectl apply -f -

# Keep envFrom pointing at auth-oauth-credentials
kubectl patch deployment auth -n puchi-backend --type json -p='[
  {"op":"replace","path":"/spec/template/spec/containers/0/envFrom","value":[{"secretRef":{"name":"auth-oauth-credentials"}}]}
]' 2>/dev/null || true

kubectl rollout restart deployment/auth -n puchi-backend
kubectl rollout status deployment/auth -n puchi-backend --timeout=90s

POD=$(kubectl get pods -n puchi-backend -l app=auth --field-selector=status.phase=Running -o jsonpath='{.items[0].metadata.name}')
echo "Pod=$POD"
kubectl exec -n puchi-backend "$POD" -- sh -c 'echo TIKTOK_KEY_len=${#TIKTOK_CLIENT_KEY}; echo GOOGLE_len=${#GOOGLE_CLIENT_ID}'
echo DONE
