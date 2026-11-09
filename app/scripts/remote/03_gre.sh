#!/usr/bin/env bash
# Configure GRE endpoints.
# On S1: ROLE=s1 S1_PUBLIC=... S2_PUBLIC=... ./03_gre.sh
# On S2: ROLE=s2 S1_PUBLIC=... S2_PUBLIC=... ./03_gre.sh
set -euo pipefail
: "${ROLE:?s1|s2}"
: "${S1_PUBLIC:?}"
: "${S2_PUBLIC:?}"
GRE_LOCAL_S1="${GRE_LOCAL_S1:-10.10.10.1}"
GRE_LOCAL_S2="${GRE_LOCAL_S2:-10.10.10.2}"

ip link del gre-vpn 2>/dev/null || true
if [[ "$ROLE" == "s1" ]]; then
  ip tunnel add gre-vpn mode gre remote "$S2_PUBLIC" local "$S1_PUBLIC" ttl 255
  ip addr add "${GRE_LOCAL_S1}/30" dev gre-vpn
else
  ip tunnel add gre-vpn mode gre remote "$S1_PUBLIC" local "$S2_PUBLIC" ttl 255
  ip addr add "${GRE_LOCAL_S2}/30" dev gre-vpn
fi
ip link set gre-vpn up
echo "[ok] GRE up on $ROLE"
ip -br addr show gre-vpn
