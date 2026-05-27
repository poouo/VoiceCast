$ErrorActionPreference = "Stop"
$env:GOPROXY = "https://goproxy.cn,direct"
$version = if ($env:VOICECAST_VERSION) { $env:VOICECAST_VERSION } else { "0.1.4" }
$build = if ($env:VOICECAST_BUILD) { $env:VOICECAST_BUILD } else { "5" }

if (-not $env:ANDROID_NDK_HOME) {
    throw "ANDROID_NDK_HOME is required. Install Android NDK first."
}

Copy-Item -Force -LiteralPath "assets\logo\voicecast-256.png" -Destination "cmd\voicecast-android\Icon.png"
$manifestPath = "cmd\voicecast-android\AndroidManifest.xml"
$manifest = Get-Content -Raw -LiteralPath $manifestPath
$manifest = $manifest -replace 'android:versionName="[^"]+"', "android:versionName=`"$version`""
$manifest = $manifest -replace 'android:versionCode="[0-9]+"', "android:versionCode=`"$build`""
Set-Content -LiteralPath $manifestPath -Value $manifest -Encoding UTF8
Push-Location "cmd\voicecast-android"
try {
    go run fyne.io/fyne/v2/cmd/fyne package `
        -os android/arm64 `
        -appID com.poouo.voicecast `
        -name VoiceCast `
        -icon Icon.png `
        -appVersion $version `
        -appBuild $build `
        -release
} finally {
    Pop-Location
}

$apk = Get-ChildItem -Path "cmd\voicecast-android" -Filter "*.apk" | Select-Object -First 1
if (-not $apk) {
    throw "APK was not generated."
}

New-Item -ItemType Directory -Force -Path "dist" | Out-Null
Move-Item -Force -LiteralPath $apk.FullName -Destination "dist\VoiceCast-Android.apk"
Write-Host "Built dist/VoiceCast-Android.apk"
