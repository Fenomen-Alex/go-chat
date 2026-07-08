param(
    [string]$Version = "v0.1.0"
)

$Repo = "Fenomen-Alex/go-chat"
$Arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64' -or $env:PROCESSOR_ARCHITEW6432 -eq 'ARM64') { "arm64" } else { "amd64" }
$Binary = "chat-windows-$Arch.exe"
$Url = "https://github.com/$Repo/releases/download/$Version/$Binary"
$ChecksumsUrl = "https://github.com/$Repo/releases/download/$Version/checksums.txt"

Write-Output "Downloading $Binary $Version..."

$OutDir = "$env:TEMP\go-chat-install"
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
$OutFile = "$OutDir\$Binary"

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
Invoke-WebRequest -Uri $Url -OutFile $OutFile -UseBasicParsing

$Checksums = Invoke-WebRequest -Uri $ChecksumsUrl -UseBasicParsing
$Expected = ($Checksums.Content -split "`n" | Where-Object { $_ -match $Binary } | ForEach-Object { $_ -split "\s+" | Select-Object -First 1 })
if (-not $Expected) {
    Write-Error "Binary not found in checksums — aborting"
    exit 1
}
$Actual = (Get-FileHash $OutFile -Algorithm SHA256).Hash.ToLower()
if ($Actual -ne $Expected) {
    Write-Error "Checksum mismatch"
    exit 1
}
Write-Output "Checksum verified"

$InstallDir = "$env:ProgramFiles\go-chat"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Move-Item -Force $OutFile "$InstallDir\chat.exe"

$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
}

Write-Output "Installed: $InstallDir\chat.exe"
Write-Output "Run: chat"
