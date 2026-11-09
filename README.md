# vpn

VPN Manager — OpenVPN control plane for a two-server GRE pair.

## App `0.1.0`

See [`app/README.md`](app/README.md).

```bash
cd app
go build -o bin/vpn-manager ./cmd/vpn-manager
./bin/vpn-manager
```

Then open http://127.0.0.1:8787/

## macOS `0.1.2`

Download from Releases:
- `vpn-manager-0.1.2-darwin-arm64.zip` (Apple Silicon)
- `vpn-manager-0.1.2-darwin-amd64.zip` (Intel)

Then run `./vpn-manager` and open http://127.0.0.1:8787/

## Android

```bash
cd android
./gradlew assembleRelease
```

Release APK: see GitHub Releases (`vpn-manager-0.1.0-android.apk`).

## UI prototype archive

Earlier HTML prototype: `interface/vpn-manager-1.html`
