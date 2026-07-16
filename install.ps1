$ErrorActionPreference = "Stop"

$Repository = "SlateNull/code-init"
$InstallDir = if ($env:CODE_INIT_INSTALL_DIR) { $env:CODE_INIT_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "Programs\code-init" }
$Version = if ($env:CODE_INIT_VERSION) { $env:CODE_INIT_VERSION } else { "latest" }

$Architecture = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { throw "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

$Asset = "code-init_windows_${Architecture}.exe"
if ($Version -eq "latest") {
    $BaseUrl = "https://github.com/$Repository/releases/latest/download"
} else {
    $Tag = if ($Version.StartsWith("v")) { $Version } else { "v$Version" }
    $BaseUrl = "https://github.com/$Repository/releases/download/$Tag"
}

$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("code-init-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $TempDir | Out-Null

try {
    $Download = Join-Path $TempDir $Asset
    $Checksums = Join-Path $TempDir "checksums.txt"

    Write-Host "Downloading $Version for windows/$Architecture..."
    curl.exe --fail --location --silent --show-error --retry 3 "$BaseUrl/$Asset" --output $Download
    if ($LASTEXITCODE -ne 0) { throw "Binary download failed; verify that the release and platform asset exist" }
    curl.exe --fail --location --silent --show-error --retry 3 "$BaseUrl/checksums.txt" --output $Checksums
    if ($LASTEXITCODE -ne 0) { throw "Checksum download failed" }

    $ChecksumLine = Get-Content $Checksums | Where-Object { $_ -match "^[0-9a-fA-F]{64}\s+\*?$([regex]::Escape($Asset))$" } | Select-Object -First 1
    if (-not $ChecksumLine) { throw "Release checksum does not include $Asset" }
    $Expected = ($ChecksumLine -split "\s+")[0].ToLowerInvariant()
    $Actual = (Get-FileHash -Algorithm SHA256 $Download).Hash.ToLowerInvariant()
    if ($Actual -ne $Expected) { throw "Checksum verification failed" }

    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    $Destination = Join-Path $InstallDir "code-init.exe"
    Move-Item -Force $Download $Destination

    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $PathEntries = if ($UserPath) { $UserPath.Split(';', [System.StringSplitOptions]::RemoveEmptyEntries) } else { @() }
    if ($PathEntries -notcontains $InstallDir) {
        $NewPath = (($PathEntries + $InstallDir) -join ';')
        [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
        Write-Host "Added $InstallDir to your user PATH. Open a new terminal before running code-init."
    }

    Write-Host "Installed code-init to $Destination"
} finally {
    Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $TempDir
}
