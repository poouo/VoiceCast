$ErrorActionPreference = "Stop"
$env:GOPROXY = "https://goproxy.cn,direct"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:GOAMD64 = "v1"
$env:CGO_ENABLED = "1"

New-Item -ItemType Directory -Force -Path "dist" | Out-Null

$windres = Get-Command windres -ErrorAction SilentlyContinue
if (-not $windres) {
    $windres = Get-Command "C:\TDM-GCC-64\bin\windres.exe" -ErrorAction SilentlyContinue
}
if (-not $windres) {
    throw "windres not found. Install MinGW-w64 or MSYS2 first."
}

$windresPath = $windres.Source
$windresArgs = @(
    "-O", "coff",
    "-i", "cmd\voicecast-windows\voicecast.rc",
    "-o", "cmd\voicecast-windows\voicecast_windows.syso"
)
& $windresPath @windresArgs

go build -trimpath -ldflags "-s -w -H=windowsgui -linkmode=internal" -o "dist/VoiceCast-Windows-amd64.exe" ./cmd/voicecast-windows
go version -m "dist/VoiceCast-Windows-amd64.exe"
Write-Host "Built dist/VoiceCast-Windows-amd64.exe"
