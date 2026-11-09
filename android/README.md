# VPN Manager Android `0.1.1`

Android app with the same OpenVPN control UI as the desktop build.

## Fixes in 0.1.1
- Working **Add profile from file** via native Android document picker
- Full-bleed phone layout (no desktop frame/zoom)
- WebChromeClient file chooser fallback

## Build

```bash
export ANDROID_HOME=$HOME/android-sdk
cd android
./gradlew assembleRelease
```

## Install

From GitHub Releases: `vpn-manager-0.1.1-android.apk`

```bash
adb install -r vpn-manager-0.1.1-android.apk
```
