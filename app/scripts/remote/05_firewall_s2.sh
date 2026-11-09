#!/usr/bin/env bash
# Run on Server 2 AFTER GRE is up (prefer via GRE / jump from S1).
# Locks admin SSH to GRE subnet; public IP must not be used for management.
set -euo pipefail
GRE_NET="${GRE_NET:-10.10.10.0/30}"
SSH_PORT="${SSH_PORT:-22}"

iptables -C INPUT -s "$GRE_NET" -p tcp --dport "$SSH_PORT" -j ACCEPT 2>/dev/null \
  || iptables -I INPUT -s "$GRE_NET" -p tcp --dport "$SSH_PORT" -j ACCEPT

# Drop new SSH from elsewhere (keep established).
iptables -C INPUT -p tcp --dport "$SSH_PORT" -m state --state ESTABLISHED,RELATED -j ACCEPT 2>/dev/null \
  || iptables -A INPUT -p tcp --dport "$SSH_PORT" -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -C INPUT -p tcp --dport "$SSH_PORT" -j DROP 2>/dev/null \
  || iptables -A INPUT -p tcp --dport "$SSH_PORT" -j DROP

echo "[ok] Server 2 SSH management restricted to GRE $GRE_NET"
echo "[warn] Do not manage Server 2 via its public IP from the app"
