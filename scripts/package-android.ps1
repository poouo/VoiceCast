$ErrorActionPreference = "Stop"
$env:GOPROXY = "https://goproxy.cn,direct"

if (-not $env:ANDROID_NDK_HOME) {
    throw "ANDROID_NDK_HOME is required. Install Android NDK first."
}

go run fyne.io/fyne/v2/cmd/fyne package `
    -os android `
    -appID com.poouo.voicecast `
    -name VoiceCast `
    -icon assets/logo/voicecast.svg `
    ./cmd/voicecast-android

$apk = Get-ChildItem -Path . -Filter "*.apk" | Select-Object -First 1
if (-not $apk) {
    throw "APK was not generated."
}

New-Item -ItemType Directory -Force -Path "dist" | Out-Null
Move-Item -Force -LiteralPath $apk.FullName -Destination "dist\VoiceCast-Android.apk"
Write-Host "Built dist/VoiceCast-Android.apk"
