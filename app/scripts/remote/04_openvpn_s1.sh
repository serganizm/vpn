#!/usr/bin/env bash
# Minimal OpenVPN server skeleton on Server 1 (entry). Clients must remote here.
set -euo pipefail
mkdir -p /etc/openvpn/server
cat >/etc/openvpn/server/server.conf <<'EOF'
port 1194
proto udp
dev tun
topology subnet
server 10.8.0.0 255.255.255.0
persist-key
persist-tun
keepalive 10 120
cipher AES-256-GCM
auth SHA256
user nobody
group nogroup
verb 3
# Push split/all routes from control plane later.
EOF
echo "[ok] OpenVPN server skeleton written on Server 1"
echo "[note] Generate PKI (ca/server/client) before enabling the service"
