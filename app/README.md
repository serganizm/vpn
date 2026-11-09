# VPN Manager `0.1.0`

Cross-platform control app (first version) for an **OpenVPN** pair:

- **Server 1** — entry (clients + admin connect here)
- **Server 2** — exit abroad (public IP may be blocked)
- After install, **Server 2 is managed only through GRE via Server 1**

## Run

```bash
cd app
go build -o bin/vpn-manager ./cmd/vpn-manager
./bin/vpn-manager
```

Open http://127.0.0.1:8787/

### Modes

| Mode | Flag / env | Behavior |
|------|------------|----------|
| dry-run (default) | `--mode dry-run` or `VPN_MANAGER_MODE=dry-run` | Simulates install steps, no SSH |
| live | `--mode live` | SSH to S1, then S1→S2 bootstrap probe |

## API (local)

- `GET /api/v1/info`
- `POST /api/v1/install` → poll `GET /api/v1/install/{id}`
- `GET /api/v1/pairs`
- `POST /api/v1/pairs/{id}/settings` (requires S1 session + GRE)
- `GET /api/v1/users/{id}/config.ovpn`

## Layout

```
app/
  cmd/vpn-manager/   # binary + embedded web UI
  internal/          # API, install orchestrator, OpenVPN config
  scripts/remote/    # bash helpers for S1/S2/GRE/OpenVPN
```

## Not in 0.1.0 yet

- Bundled OpenVPN data-plane inside Android/desktop VPN stack
- Full PKI generation and package install on remote hosts
- Production hardening of SSH host keys
