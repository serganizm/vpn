#!/usr/bin/env bash
# Run on Server 1. Bootstraps Server 2 over SSH (public IP only used here).
# Usage: S2_HOST=x.x.x.x S2_USER=root S2_PORT=22 ./02_bootstrap_s2_via_s1.sh
set -euo pipefail
: "${S2_HOST:?set S2_HOST}"
: "${S2_USER:=root}"
: "${S2_PORT:=22}"

ssh -p "$S2_PORT" -o StrictHostKeyChecking=accept-new "${S2_USER}@${S2_HOST}" 'bash -s' <<'EOS'
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install -y openvpn iptables iproute2
modprobe gre || true
echo "[ok] Server 2 packages ready (via S1 jump)"
EOS
