# Changelog

## 0.1.0 — 2026-11-09

First application release.

- Single Go binary: embedded UI + local API (`vpn-manager`)
- OpenVPN only
- Install flow: PC → Server 1 → Server 2 (bootstrap), then Server 2 admin via GRE only
- `.ovpn` always uses `remote <Server1>`; Server 2 public IP is forbidden in client configs
- Dry-run mode by default; `--mode live` enables real SSH probes
- Remote helper scripts under `scripts/remote/`
