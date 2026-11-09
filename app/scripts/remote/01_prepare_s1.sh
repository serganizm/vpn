#!/usr/bin/env bash
# Run on Server 1. Prepares packages for OpenVPN + GRE.
set -euo pipefail
export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install -y openvpn easy-rsa iptables iproute2 curl
modprobe gre || true
echo "[ok] Server 1 packages ready"
