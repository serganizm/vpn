# Remote install scripts (OpenVPN + GRE)

These scripts are intended to run **on Server 1** during bootstrap.

## Rules

1. Admin PC connects **only to Server 1**.
2. Server 2 is reached as `S1 → SSH → S2` once (bootstrap), then **only via GRE**.
3. OpenVPN clients always use `remote <Server1> 1194`.
4. After GRE is up, public SSH management on Server 2 should be closed.

## Order

1. `01_prepare_s1.sh`
2. `02_bootstrap_s2_via_s1.sh` (from S1)
3. `03_gre.sh` (both sides)
4. `04_openvpn_s1.sh`
5. `05_firewall_s2.sh` (on S2 via GRE / jump)
