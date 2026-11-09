# Release keystore

Generate locally under `android/keystore/` (gitignored):

```bash
keytool -genkeypair -v \
  -keystore keystore/vpn-manager-release.jks \
  -alias vpnmanager \
  -keyalg RSA -keysize 2048 -validity 10000 \
  -storepass vpnmanager0 \
  -keypass vpnmanager0 \
  -dname "CN=VPN Manager, OU=Dev, O=serganizm, L=Internet, ST=NA, C=RU"
```

Demo passwords are for v0.1 only. Rotate for production.
