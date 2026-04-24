[CmdletBinding()]
param(
    [string]$Repo = 'gqcn/linapro',
    [string]$Ref = '',
    [string]$Dir = '',
    [string]$Name = '',
    [switch]$CurrentDir,
    [switch]$Force,
    [switch]$Help
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$DefaultFallbackRef = 'main'
$Script:DownloadedFrom = ''
$Script:WorkDir = $null
$Script:ResolvedRefSource = ''
$Script:RefExplicit = $PSBoundParameters.ContainsKey('Ref') -and -not [string]::IsNullOrWhiteSpace($Ref)

function Show-Usage {
    @"
LinaPro installer for Windows PowerShell.

Usage:
  install.ps1 [-Repo <owner/name>] [-Ref <value>] [-Dir <path>] [-Name <directory>] [-CurrentDir] [-Force] [-Help]

Parameters:
  -Repo         GitHub repository to download. Default: gqcn/linapro
  -Ref          Branch, tag, or commit reference. Default: latest stable tag
                Fallback: main when no stable tag is available.
  -Dir          Install into the specified directory.
  -Name         Install into a new child directory under the current path.
  -CurrentDir   Install directly into the current working directory.
  -Force        Allow overlay install into a non-empty target directory.
  -Help         Show this help message.

Examples:
  .\install.ps1
  .\install.ps1 -Ref v0.1.0 -Name linapro-v0.1.0
  .\install.ps1 -Dir C:\Workspace\linapro
  .\install.ps1 -CurrentDir -Force

Advanced environment variables:
  LINAPRO_INSTALL_ARCHIVE_PATH  Use a local .zip archive instead of downloading.
  LINAPRO_INSTALL_STABLE_REF    Override the auto-detected stable tag.
"@
}

function Normalize-Repo {
    param([string]$InputRepo)

    $repo = $InputRepo.Trim()

    if ($repo -match '^https?://github\.com/([^/]+)/([^/]+)/?$') {
        $repo = "$($Matches[1])/$($Matches[2])"
    } elseif ($repo -match '^https?://github\.com/([^/]+)/([^/]+?)\.git/?$') {
        $repo = "$($Matches[1])/$($Matches[2])"
    } elseif ($repo -match '^git@github\.com:([^/]+)/([^/]+?)(\.git)?$') {
        $repo = "$($Matches[1])/$($Matches[2])"
    }

    if ($repo.EndsWith('.git')) {
        $repo = $repo.Substring(0, $repo.Length - 4)
    }

    if ($repo -notmatch '^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$') {
        throw "Unsupported repository value '$InputRepo'. Use owner/name or a GitHub URL."
    }

    return $repo
}

function Get-ArchiveCandidates {
    param(
        [string]$NormalizedRepo,
        [string]$RefValue
    )

    return @(
        "https://codeload.github.com/$NormalizedRepo/zip/refs/heads/$RefValue",
        "https://codeload.github.com/$NormalizedRepo/zip/refs/tags/$RefValue",
        "https://codeload.github.com/$NormalizedRepo/zip/$RefValue"
    )
}

function Resolve-EffectiveRef {
    param([string]$NormalizedRepo)

    if ($Script:RefExplicit) {
        $Script:ResolvedRefSource = 'user provided'
        return $Ref
    }

    if ($env:LINAPRO_INSTALL_STABLE_REF) {
        $Script:ResolvedRefSource = 'stable override'
        return $env:LINAPRO_INSTALL_STABLE_REF
    }

    try {
        $tagResponse = Invoke-RestMethod -Uri "https://api.github.com/repos/$NormalizedRepo/tags?per_page=100"
        $stableTags = @($tagResponse | Where-Object {
                $_.name -match '^[vV]?\d+\.\d+\.\d+$'
            })

        if ($stableTags.Count -gt 0) {
            $latestStableTag = $stableTags |
                Sort-Object -Property @{ Expression = { [version]($_.name.TrimStart('v', 'V')) } } |
                Select-Object -Last 1

            if ($latestStableTag -and $latestStableTag.name) {
                $Script:ResolvedRefSource = 'latest stable tag'
                return $latestStableTag.name
            }
        }
    } catch {
        # Fall back to the default branch when GitHub tag discovery is unavailable.
    }

    $Script:ResolvedRefSource = 'fallback branch'
    return $DefaultFallbackRef
}

function Test-DirectoryHasEntries {
    param([string]$Path)

    if (-not (Test-Path -LiteralPath $Path -PathType Container)) {
        return $false
    }

    return (Get-ChildItem -LiteralPath $Path -Force | Measure-Object).Count -gt 0
}

function Resolve-TargetDir {
    param([string]$NormalizedRepo)

    if ($CurrentDir) {
        return (Get-Location).ProviderPath
    }

    if ($Dir) {
        return [System.IO.Path]::GetFullPath($Dir)
    }

    $childName = $Name
    if (-not $childName) {
        $childName = Split-Path -Path $NormalizedRepo -Leaf
    }

    return [System.IO.Path]::GetFullPath((Join-Path -Path (Get-Location).ProviderPath -ChildPath $childName))
}

function Save-Archive {
    param(
        [string]$NormalizedRepo,
        [string]$RefValue,
        [string]$ArchiveFile
    )

    if ($env:LINAPRO_INSTALL_ARCHIVE_PATH) {
        if (-not (Test-Path -LiteralPath $env:LINAPRO_INSTALL_ARCHIVE_PATH -PathType Leaf)) {
            throw "LINAPRO_INSTALL_ARCHIVE_PATH points to a missing file: $($env:LINAPRO_INSTALL_ARCHIVE_PATH)"
        }

        Copy-Item -LiteralPath $env:LINAPRO_INSTALL_ARCHIVE_PATH -Destination $ArchiveFile -Force
        $Script:DownloadedFrom = "local archive: $($env:LINAPRO_INSTALL_ARCHIVE_PATH)"
        return
    }

    foreach ($candidate in Get-ArchiveCandidates -NormalizedRepo $NormalizedRepo -RefValue $RefValue) {
        try {
            Invoke-WebRequest -Uri $candidate -OutFile $ArchiveFile | Out-Null
            $Script:DownloadedFrom = $candidate
            return
        } catch {
            Remove-Item -LiteralPath $ArchiveFile -Force -ErrorAction SilentlyContinue
        }
    }

    throw "Failed to download archive for repository '$NormalizedRepo' and ref '$RefValue'."
}

function Expand-SourceArchive {
    param(
        [string]$ArchiveFile,
        [string]$ExtractRoot
    )

    New-Item -ItemType Directory -Path $ExtractRoot -Force | Out-Null
    Expand-Archive -LiteralPath $ArchiveFile -DestinationPath $ExtractRoot -Force

    $children = Get-ChildItem -LiteralPath $ExtractRoot -Force | Where-Object { $_.PSIsContainer }
    if ($children.Count -ne 1) {
        throw 'Expected the archive to extract into a single top-level directory.'
    }

    return $children[0].FullName
}

function Copy-SourceContents {
    param(
        [string]$SourceDir,
        [string]$TargetDir
    )

    New-Item -ItemType Directory -Path $TargetDir -Force | Out-Null
    Get-ChildItem -LiteralPath $SourceDir -Force | ForEach-Object {
        Copy-Item -LiteralPath $_.FullName -Destination $TargetDir -Recurse -Force
    }
}

function Write-CheckLine {
    param(
        [string]$Label,
        [string]$CommandName,
        [string[]]$Arguments
    )

    $command = Get-Command -Name $CommandName -ErrorAction SilentlyContinue
    if ($null -eq $command) {
        Write-Host "  [MISSING] $Label"
        return $false
    }

    try {
        $output = & $CommandName @Arguments 2>$null | Select-Object -First 1
        if ([string]::IsNullOrWhiteSpace([string]$output)) {
            $output = 'detected'
        }
        Write-Host "  [OK] $Label: $output"
        return $true
    } catch {
        Write-Host "  [OK] $Label: detected"
        return $true
    }
}

function Run-EnvironmentCheck {
    $missingCount = 0

    Write-Host ''
    Write-Host 'Environment check:'
    if (-not (Write-CheckLine -Label 'Go' -CommandName 'go' -Arguments @('version'))) { $missingCount++ }
    if (-not (Write-CheckLine -Label 'Node.js' -CommandName 'node' -Arguments @('--version'))) { $missingCount++ }
    if (-not (Write-CheckLine -Label 'pnpm' -CommandName 'pnpm' -Arguments @('--version'))) { $missingCount++ }
    if (-not (Write-CheckLine -Label 'MySQL' -CommandName 'mysql' -Arguments @('--version'))) { $missingCount++ }
    if (-not (Write-CheckLine -Label 'make' -CommandName 'make' -Arguments @('--version'))) { $missingCount++ }

    if ($missingCount -gt 0) {
        Write-Host ''
        Write-Host 'Some dependencies are missing. Install them before running the LinaPro bootstrap commands.'
    }
}

function Write-NextSteps {
    param([string]$TargetDir)

    Write-Host ''
    Write-Host "Project directory: $TargetDir"
    Write-Host "Archive source: $Script:DownloadedFrom"
    Write-Host ''
    Write-Host 'Next steps:'
    Write-Host "  1. Set-Location '$TargetDir'"
    Write-Host '  2. make init confirm=init'
    Write-Host '  3. make mock confirm=mock'
    Write-Host '  4. make dev'
    Write-Host ''
    Write-Host 'The installer only bootstraps the source tree and environment check.'
}

function Invoke-Cleanup {
    if ($Script:WorkDir -and (Test-Path -LiteralPath $Script:WorkDir)) {
        Remove-Item -LiteralPath $Script:WorkDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

try {
    if ($Help) {
        Show-Usage
        exit 0
    }

    if ($CurrentDir -and $Dir) {
        throw '-CurrentDir and -Dir cannot be used together.'
    }

    if ($CurrentDir -and $Name) {
        throw '-CurrentDir and -Name cannot be used together.'
    }

    if ($Dir -and $Name) {
        throw '-Dir and -Name cannot be used together.'
    }

    $normalizedRepo = Normalize-Repo -InputRepo $Repo
    $resolvedRef = Resolve-EffectiveRef -NormalizedRepo $normalizedRepo
    $targetDir = Resolve-TargetDir -NormalizedRepo $normalizedRepo

    Write-Host "Repository: $normalizedRepo"
    Write-Host "Resolved ref: $resolvedRef [$Script:ResolvedRefSource]"
    Write-Host "Target directory: $targetDir"

    if ((Test-Path -LiteralPath $targetDir -PathType Container) -and (Test-DirectoryHasEntries -Path $targetDir) -and (-not $Force)) {
        throw "Target directory '$targetDir' is not empty; rerun with -Force to overlay the source tree."
    }

    $Script:WorkDir = Join-Path -Path ([System.IO.Path]::GetTempPath()) -ChildPath ([System.Guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $Script:WorkDir -Force | Out-Null

    $archiveFile = Join-Path -Path $Script:WorkDir -ChildPath 'source.zip'
    $extractRoot = Join-Path -Path $Script:WorkDir -ChildPath 'extract'

    Save-Archive -NormalizedRepo $normalizedRepo -RefValue $resolvedRef -ArchiveFile $archiveFile
    $sourceDir = Expand-SourceArchive -ArchiveFile $archiveFile -ExtractRoot $extractRoot
    Copy-SourceContents -SourceDir $sourceDir -TargetDir $targetDir

    Write-Host 'LinaPro source bootstrapped successfully.'
    Run-EnvironmentCheck
    Write-NextSteps -TargetDir $targetDir
} catch {
    Write-Error $_
    exit 1
} finally {
    Invoke-Cleanup
}
