# VPN Manager Android `0.1.0`

Android app with the same OpenVPN control UI as the desktop build.

- Clients connect to **Server 1**
- **Server 2** is managed only via **GRE through Server 1**
- In-app dry-run API (no separate backend process)
- Embedded Web UI + native bridge

## Build release APK

```bash
export ANDROID_HOME=$HOME/android-sdk
cd android
./gradlew assembleRelease
```

APK: `app/build/outputs/apk/release/app-release-unsigned.apk`  
(or signed `app-release.apk` when a keystore is configured)

## Install

```bash
adb install -r app/build/outputs/apk/release/*.apk
```


## Release artifact

`dist/vpn-manager-0.1.0-android.apk`
