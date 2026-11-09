# Changelog

## 0.1.1-android — 2026-11-09

- Fix Add profile from file on Android (native document picker)
- Full-bleed mobile layout without desktop phone-frame zoom
- WebChromeClient file chooser support

# Changelog

## 0.1.0 — 2026-11-09

First application release.

- Single Go binary: embedded UI + local API (`vpn-manager`)
- OpenVPN only
- Install flow: PC → Server 1 → Server 2 (bootstrap), then Server 2 admin via GRE only
- `.ovpn` always uses `remote <Server1>`; Server 2 public IP is forbidden in client configs
- Dry-run mode by default; `--mode live` enables real SSH probes
- Remote helper scripts under `scripts/remote/`


## 0.1.0-android — 2026-11-09

- Android APK with the same OpenVPN control UI
- In-app dry-run API (S1 bootstrap, S2 via GRE)
- Package: `com.serganizm.vpnmanager`
