param(
  [string]$Version = $env:RUNECTX_VERSION,
  [string]$InstallDir = $(if ($env:RUNECTX_INSTALL_DIR) { $env:RUNECTX_INSTALL_DIR } else { Join-Path $HOME ".local/bin" }),
  [switch]$Yes,
  [switch]$Help
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Show-Usage {
  @"
Install runectx from GitHub Releases.

Usage:
  install-runectx.ps1 [-Version TAG] [-InstallDir DIR] [-Yes]

Options:
  -Version TAG      Install a specific release tag (for example v0.1.0-alpha.8)
                    Defaults to the latest published release.
  -InstallDir DIR   Install directory for runectx (default: `$HOME/.local/bin)
  -Yes              Skip confirmation prompt and continue install.
  -Help             Show this help text.

Environment:
  RUNECTX_VERSION      Same as -Version
  RUNECTX_INSTALL_DIR  Same as -InstallDir
"@
}

function Resolve-LatestTag {
  $response = Invoke-WebRequest -UseBasicParsing -Uri "https://github.com/runecode-systems/runecontext/releases/latest"
  $segments = $response.BaseResponse.ResponseUri.AbsolutePath.Split('/', [System.StringSplitOptions]::RemoveEmptyEntries)
  if ($segments.Length -lt 1) {
    throw "failed to resolve latest release tag"
  }
  return $segments[$segments.Length - 1]
}

function Resolve-Arch {
  switch ($env:PROCESSOR_ARCHITECTURE.ToLowerInvariant()) {
    "amd64" { return "amd64" }
    "x86" { throw "x86 is not supported" }
    "arm64" { return "arm64" }
    default {
      if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant() -eq "x64") {
        return "amd64"
      }
      throw "unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
    }
  }
}

function Test-VersionTag {
  param(
    [string]$Tag
  )

  if ($Tag -match '^v[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.-]+)?$') {
    return $true
  }

  return $false
}

function Parse-Sha256Entry {
  param(
    [string]$ChecksumsPath,
    [string[]]$CandidateArchives
  )

  foreach ($candidate in $CandidateArchives) {
    $pattern = "\s{2}" + [regex]::Escape($candidate) + '$'
    $entry = Select-String -Path $ChecksumsPath -Pattern $pattern | Select-Object -First 1
    if ($entry) {
      $parts = ($entry.Line -split '\s+', 2)
      if ($parts.Count -ne 2 -or $parts[1] -ne $candidate) {
        throw "SHA256SUMS entry is malformed for $candidate"
      }
      return @{ Archive = $candidate; Hash = $parts[0].ToLowerInvariant() }
    }
  }

  return $null
}

if ($Help) {
  Show-Usage
  exit 0
}

$repo = "runecode-systems/runecontext"
if ([string]::IsNullOrWhiteSpace($Version)) {
  $Version = Resolve-LatestTag
}

if (-not (Test-VersionTag -Tag $Version)) {
  throw "invalid release tag '$Version' (expected format like v0.1.0-alpha.8)"
}

$arch = Resolve-Arch
$checksums = "SHA256SUMS"
$baseUrl = "https://github.com/$repo/releases/download/$Version"

$candidateArchives = @(
  "runecontext_${Version}_windows_${arch}.zip",
  "runecontext_${Version}_windows_${arch}.tar.gz"
)

$workDir = Join-Path $env:TEMP ("runectx-install-" + [guid]::NewGuid())
$null = New-Item -ItemType Directory -Force -Path $workDir

try {
  Write-Host "Resolving release: $Version"
  Write-Host "Target platform: windows/$arch"
  Write-Host "Install destination: $(Join-Path $InstallDir 'runectx.exe')"

  $checksumsPath = Join-Path $workDir $checksums
  Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/$checksums" -OutFile $checksumsPath

  $shaEntry = Parse-Sha256Entry -ChecksumsPath $checksumsPath -CandidateArchives $candidateArchives
  if (-not $shaEntry) {
    throw "quick-install binary archive is not published for windows/$arch in release $Version. Use the manual repo-bundle flow in docs/install-verify.md or build locally with just build."
  }

  $archive = $shaEntry.Archive
  $archivePath = Join-Path $workDir $archive
  Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/$archive" -OutFile $archivePath

  $actualHash = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()

  Write-Host ""
  Write-Host "Checksum verification:"
  Write-Host "  archive:  $archive"
  Write-Host "  expected: $($shaEntry.Hash)"
  Write-Host "  actual:   $actualHash"

  if ($actualHash -ne $shaEntry.Hash) {
    throw "checksum verification failed"
  }

  Write-Host "  result:   OK"
  Write-Host ""

  if (-not $Yes) {
    $reply = Read-Host "Continue with installation? [y/N]"
    if ([string]::IsNullOrWhiteSpace($reply) -or $reply -notmatch '^(?i:y|yes)$') {
      Write-Host "Installation cancelled."
      exit 0
    }
  }

  $extractDir = Join-Path $workDir "unpack"
  $null = New-Item -ItemType Directory -Force -Path $extractDir

  if ($archive.EndsWith(".zip")) {
    Expand-Archive -Path $archivePath -DestinationPath $extractDir -Force
  } elseif ($archive.EndsWith(".tar.gz")) {
    tar -xzf $archivePath -C $extractDir
  } else {
    throw "unsupported archive format: $archive"
  }

  $packageDir = Join-Path $extractDir ("runecontext_${Version}_windows_${arch}")
  $binDir = Join-Path $packageDir "bin"

  $candidateBinaries = @(
    Join-Path $binDir "runectx.exe",
    Join-Path $binDir "runectx"
  )

  $binaryPath = $candidateBinaries | Where-Object { Test-Path $_ } | Select-Object -First 1
  if (-not $binaryPath) {
    throw "expected runectx binary not found under $binDir"
  }

  $null = New-Item -ItemType Directory -Force -Path $InstallDir
  $installTarget = Join-Path $InstallDir "runectx.exe"
  Copy-Item -Path $binaryPath -Destination $installTarget -Force

  Write-Host ""
  Write-Host "Installed runectx to $installTarget"
  & $installTarget version

  Write-Host ""
  Write-Host "Next steps:"
  Write-Host "- Ensure your install directory is on PATH."
  Write-Host "- Run: runectx doctor --path /path/to/project"
  Write-Host "- Initialize a project: runectx init --path /path/to/project"
  Write-Host "- Sync adapter files: runectx adapter sync --path /path/to/project <tool>"
  Write-Host "- Preview upgrades: runectx upgrade --path /path/to/project"
}
finally {
  Remove-Item -Recurse -Force $workDir -ErrorAction SilentlyContinue
}
