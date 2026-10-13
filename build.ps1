<#
.SYNOPSIS
    Builds a release-ready Name It executable.

.DESCRIPTION
    Produces a Windows binary with Go symbols and DWARF data stripped and local
    paths trimmed. UPX compression is optional and disabled by default; pass
    -Upx to pack the binary with UPX using its strongest practical settings
    (--best + LZMA). The resulting executable is copied to -OutDir using a
    versioned file name.

.PARAMETER OutDir
    Directory that receives the final artifact. Defaults to "dist".

.PARAMETER Upx
    Compress the binary with UPX after stripping it.

.EXAMPLE
    ./build.ps1
    ./build.ps1 -OutDir dist -Upx
#>

[CmdletBinding()]
param(
    [string]$OutDir = "dist",
    [switch]$Upx
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

function Assert-Command {
    param([string]$Name)

    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "$Name was not found in PATH. Install it before building."
    }
}

Assert-Command wails
Assert-Command go

$versionFile = Join-Path $PSScriptRoot "internal\version\version.go"
$versionMatch = Select-String -Path $versionFile -Pattern 'Version\s*=\s*"([^"]+)"'
if (-not $versionMatch) {
    throw "Could not read the version from $versionFile."
}
$version = $versionMatch.Matches[0].Groups[1].Value

Write-Host "Building Name It v$version..."

# -clean wipes stale artifacts, -trimpath removes local paths, and the ldflags
# drop the symbol table and DWARF sections that only help while debugging.
wails build -clean -trimpath -ldflags "-s -w"
if ($LASTEXITCODE -ne 0) {
    throw "wails build failed with exit code $LASTEXITCODE."
}

$exe = Get-ChildItem -Path (Join-Path $PSScriptRoot "build\bin") -Filter "*.exe" |
    Select-Object -First 1

if (-not $exe) {
    throw "The build did not produce an executable in build\bin."
}

if (-not $Upx) {
    Write-Host "Skipping UPX compression (pass -Upx to enable it)." -ForegroundColor Yellow
}
elseif (Get-Command upx -ErrorAction SilentlyContinue) {
    Write-Host "Compressing with UPX (--best --lzma)..."
    upx --best --lzma $exe.FullName
    if ($LASTEXITCODE -ne 0) {
        throw "upx failed with exit code $LASTEXITCODE."
    }
}
else {
    Write-Warning "upx was not found; the binary is left uncompressed."
}

$outPath = Join-Path $PSScriptRoot $OutDir
New-Item -ItemType Directory -Force -Path $outPath | Out-Null

$target = Join-Path $outPath "name-it-$version-windows-amd64.exe"
Copy-Item $exe.FullName $target -Force

$size = [math]::Round((Get-Item $target).Length / 1MB, 2)
Write-Host "Artifact: $target ($size MB)" -ForegroundColor Green
