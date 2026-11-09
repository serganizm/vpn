#!/usr/bin/env bash
# Build a double-clickable VPN Manager.app (no installer) for macOS.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VERSION="$(tr -d '[:space:]' < "$ROOT/VERSION")"
ARCH="${1:-arm64}"
OUT_ROOT="${2:-$ROOT/../dist/macos}"
APP_NAME="VPN Manager"
BUNDLE_ID="com.serganizm.vpnmanager"

WORKDIR="$OUT_ROOT/build-darwin-$ARCH"
APP="$WORKDIR/$APP_NAME.app"
ZIP="$OUT_ROOT/VPN-Manager-$VERSION-macos-$ARCH.app.zip"

rm -rf "$WORKDIR"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"

echo "Building darwin/$ARCH..."
# Binary name = CFBundleExecutable so double-click runs it directly (no installer, no Terminal wrapper).
CGO_ENABLED=0 GOOS=darwin GOARCH="$ARCH" go build -C "$ROOT" -ldflags "-s -w" \
  -o "$APP/Contents/MacOS/VPNManager" ./cmd/vpn-manager
chmod +x "$APP/Contents/MacOS/VPNManager"

cat > "$APP/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key><string>VPN Manager</string>
  <key>CFBundleDisplayName</key><string>VPN Manager</string>
  <key>CFBundleIdentifier</key><string>${BUNDLE_ID}</string>
  <key>CFBundleVersion</key><string>${VERSION}</string>
  <key>CFBundleShortVersionString</key><string>${VERSION}</string>
  <key>CFBundleExecutable</key><string>VPNManager</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleSignature</key><string>????</string>
  <key>LSMinimumSystemVersion</key><string>11.0</string>
  <key>NSHighResolutionCapable</key><true/>
  <key>NSSupportsAutomaticTermination</key><false/>
  <key>NSSupportsSuddenTermination</key><false/>
  <key>LSBackgroundOnly</key><false/>
  <key>CFBundleInfoDictionaryVersion</key><string>6.0</string>
</dict>
</plist>
PLIST

printf 'APPL????' > "$APP/Contents/PkgInfo"

cat > "$OUT_ROOT/HOW-TO-RUN-MAC.txt" <<EOF
VPN Manager ${VERSION} — portable macOS .app (no installer)

1. Unzip the archive
2. Double-click "VPN Manager.app"
3. If macOS blocks it (unidentified developer):
   Right-click the app → Open → Open
   or:
   xattr -dr com.apple.quarantine "/path/to/VPN Manager.app"
4. Browser opens automatically at http://127.0.0.1:8787/

Apple Silicon → macos-arm64
Intel → macos-amd64
EOF

(
  cd "$WORKDIR"
  zip -yr "$ZIP" "$APP_NAME.app"
)

echo "Created $ZIP"
ls -lh "$ZIP"
