# LinaPro Development Script (Windows PowerShell)
# Helps Windows developers run daily tasks without make/Git Bash.
# CI-specific commands (test-scripts, etc.) remain with the Makefile system.
#
# First-time setup:
#   .\dev.ps1 check                          Verify environment
#   docker run ... postgres:14-alpine         Start PostgreSQL (if not running)
#   .\dev.ps1 init -Confirm init             Initialize database tables
#   .\dev.ps1 mock -Confirm mock             Load demo data
#
# Daily development:
#   .\dev.ps1 dev                            Start full-stack dev server
#   .\dev.ps1 status                         Check running status
#   .\dev.ps1 stop                           Stop all services
#
# Other common tasks:
#   .\dev.ps1 build                          Production build
#   .\dev.ps1 wasm -Plugin <plugin-id>            Build a dynamic plugin
#   .\dev.ps1 test-go                        Run Go unit tests
#   .\dev.ps1 help                           Show all commands

param(
    [Parameter(Position = 0)]
    [ValidateSet(
        'dev', 'stop', 'status', 'build', 'wasm', 'init', 'mock',
        'test', 'test-go',
        'check-i18n', 'check-i18n-messages',
        'image', 'image-build',
        'dao', 'ctrl', 'service', 'enums', 'pb', 'pbentity',
        'help', 'clean', 'check'
    )]
    [string]$Command,

    [string]$Platforms,
    [string]$CgoEnabled,
    [string]$OutputDir,
    [string]$BinaryName,
    [string]$Config,
    [switch]$VerboseOutput,
    [switch]$Background,

    [string]$Confirm,
    [string]$Rebuild,

    [string]$ImageName,
    [string]$ImageTag,
    [string]$Registry,
    [string]$Push,
    [string]$BaseImage,

    [string]$Plugin,

    [switch]$Help
)

$ErrorActionPreference = 'Stop'

# ============================================================
# Configuration
# ============================================================
$Script:RootDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$Script:BackendDir = Join-Path $Script:RootDir 'apps\lina-core'
$Script:FrontendDir = Join-Path $Script:RootDir 'apps\lina-vben'
$Script:TempDir = Join-Path $Script:RootDir 'temp'
$Script:PidDir = Join-Path $Script:TempDir 'pids'
$Script:BackendPidFile = Join-Path $Script:PidDir 'backend.pid'
$Script:FrontendPidFile = Join-Path $Script:PidDir 'frontend.pid'
$Script:BackendPort = 8080
$Script:FrontendPort = 5666
$Script:BackendLog = Join-Path $Script:TempDir 'lina-core.log'
$Script:FrontendLog = Join-Path $Script:TempDir 'lina-vben.log'
$Script:EmbedDir = Join-Path $Script:BackendDir 'internal\packed\public'
$Script:OutputDirDefault = Join-Path $Script:TempDir 'output'

# ============================================================
# Utility Functions
# ============================================================

function Write-Banner {
    Write-Host ''
    Write-Host '================================================' -ForegroundColor Cyan
    Write-Host '       LinaPro Framework' -ForegroundColor Cyan
    Write-Host '================================================' -ForegroundColor Cyan
    Write-Host ''
}

function Write-Step {
    param([string]$Message)
    Write-Host ">>> $Message" -ForegroundColor Yellow
}

function Write-Success {
    param([string]$Message)
    Write-Host "  [OK] $Message" -ForegroundColor Green
}

function Write-Info {
    param([string]$Message)
    Write-Host "    $Message" -ForegroundColor DarkGray
}

function Ensure-Dir {
    param([string]$Path)
    if (!(Test-Path $Path)) {
        New-Item -ItemType Directory -Force -Path $Path | Out-Null
    }
}

function Test-PortInUse {
    param([int]$Port)
    $conn = netstat -ano 2>$null | Select-String ":$Port " | Select-String 'LISTENING|ESTABLISHED'
    return ($null -ne $conn)
}

function Get-PortPids {
    param([int]$Port)
    $lines = netstat -ano 2>$null | Select-String ":$Port " | Select-String 'LISTENING'
    if ($null -eq $lines) { return @() }
    $pids = @()
    foreach ($line in $lines) {
        $trimmed = $line.ToString().Trim()
        $parts = $trimmed -split '\s+'
        $pidStr = $parts[-1]
        if ($pidStr -match '^\d+$') {
            $pids += [int]$pidStr
        }
    }
    return $pids
}

function Stop-ProcessTree {
    param([int]$Pid)
    try {
        $proc = Get-Process -Id $Pid -ErrorAction SilentlyContinue
        if ($null -ne $proc) {
            $children = Get-CimInstance -ClassName Win32_Process -Filter "ParentProcessId=$Pid" -ErrorAction SilentlyContinue
            if ($null -ne $children) {
                foreach ($child in $children) {
                    Stop-ProcessTree -Pid $child.ProcessId
                }
            }
            Stop-Process -Id $Pid -Force -ErrorAction SilentlyContinue
        }
    }
    catch { }
}

function Stop-Service {
    param(
        [string]$Name,
        [string]$PidFile,
        [int]$Port
    )
    $stopped = $false

    if (Test-Path $PidFile) {
        try {
            $pidContent = Get-Content $PidFile -Raw
            if ($pidContent -match '\d+') {
                $filePid = [int]$Matches[0]
                $proc = Get-Process -Id $filePid -ErrorAction SilentlyContinue
                if ($null -ne $proc) {
                    Stop-ProcessTree -Pid $filePid
                    $stopped = $true
                }
            }
        }
        catch { }
        Remove-Item $PidFile -Force -ErrorAction SilentlyContinue
    }

    $portPids = Get-PortPids -Port $Port
    if ($portPids.Count -gt 0) {
        foreach ($p in $portPids) {
            Stop-Process -Id $p -Force -ErrorAction SilentlyContinue
        }
        Start-Sleep -Milliseconds 500
        $portPids = Get-PortPids -Port $Port
        foreach ($p in $portPids) {
            Stop-Process -Id $p -Force -ErrorAction SilentlyContinue
        }
        $stopped = $true
    }

    if ($stopped) {
        Write-Success "$Name stopped"
    }
    else {
        Write-Info "$Name is not running"
    }
}

function Wait-HttpReady {
    param(
        [string]$Name,
        [string]$PidFile,
        [string]$Url,
        [int]$TimeoutSeconds,
        [string]$LogFile
    )
    $sw = [System.Diagnostics.Stopwatch]::StartNew()

    while ($sw.Elapsed.TotalSeconds -lt $TimeoutSeconds) {
        if (!(Test-Path $PidFile)) {
            Write-Host "  [FAIL] $Name startup failed: PID file does not exist" -ForegroundColor Red
            Write-Info "Check log: $LogFile"
            throw "$Name startup failed"
        }

        try {
            $pidContent = Get-Content $PidFile -Raw -ErrorAction Stop
            if ($pidContent -match '\d+') {
                $filePid = [int]$Matches[0]
                $proc = Get-Process -Id $filePid -ErrorAction SilentlyContinue
                if ($null -eq $proc) {
                    Write-Host "  [FAIL] $Name startup failed: process exited" -ForegroundColor Red
                    Write-Info "Check log: $LogFile"
                    throw "$Name process exited"
                }
            }
        }
        catch [System.Management.Automation.ItemNotFoundException] { }
        catch [System.IO.IOException] { }

        try {
            $response = Invoke-WebRequest -Uri $Url -TimeoutSec 2 -UseBasicParsing -ErrorAction SilentlyContinue
            if ($null -ne $response) {
                Write-Success "$Name is ready: $Url"
                return
            }
        }
        catch { }

        Start-Sleep -Seconds 1
    }

    Write-Host "  [FAIL] $Name startup timed out ($($TimeoutSeconds)s): $Url" -ForegroundColor Red
    Write-Info "Check log: $LogFile"
    $errLog = $LogFile -replace '\.log$', '.err.log'
    if (Test-Path $errLog) {
        Write-Host '  Last 10 lines of error log:' -ForegroundColor DarkGray
        Get-Content $errLog -Tail 10 | ForEach-Object { Write-Host "    $_" -ForegroundColor DarkGray }
    }
    throw "$Name startup timed out"
}

function Get-BuildEnv {
    $argsList = @()
    if ($Platforms) { $argsList += "--platforms=$Platforms" }
    if ($CgoEnabled) { $argsList += "--cgo-enabled=$CgoEnabled" }
    if ($OutputDir) { $argsList += "--output-dir=$OutputDir" }
    if ($BinaryName) { $argsList += "--binary-name=$BinaryName" }
    if ($Config) { $argsList += "--config=$Config" }

    $output = & go run ./hack/tools/image-builder --print-build-env $argsList 2>&1
    if ($LASTEXITCODE -ne 0) {
        throw 'Failed to read build environment'
    }

    $envVars = @{}
    foreach ($line in $output) {
        if ($line -match '^(\w+)=(.*)$') {
            $val = $Matches[2]
            if ($val.StartsWith('"') -and $val.EndsWith('"')) {
                $val = $val.Substring(1, $val.Length - 2)
            }
            $envVars[$Matches[1]] = $val
        }
    }
    return $envVars
}

function Get-ImageBuilderArgs {
    $ibArgs = @()
    if ($ImageName) { $ibArgs += "--image=$ImageName" }
    if ($ImageTag) { $ibArgs += "--tag=$ImageTag" }
    if ($Registry) { $ibArgs += "--registry=$Registry" }
    if ($Push) { $ibArgs += "--push=$Push" }
    if ($Platforms) { $ibArgs += "--platforms=$Platforms" }
    if ($CgoEnabled) { $ibArgs += "--cgo-enabled=$CgoEnabled" }
    if ($OutputDir) { $ibArgs += "--output-dir=$OutputDir" }
    if ($BinaryName) { $ibArgs += "--binary-name=$BinaryName" }
    if ($BaseImage) { $ibArgs += "--base-image=$BaseImage" }
    if ($Config) { $ibArgs += "--config=$Config" }
    if ($VerboseOutput) { $ibArgs += '--verbose=1' }
    return $ibArgs
}

# ============================================================
# Internal: prepare-packed-assets (replaces hack/scripts/prepare-packed-assets.sh)
# ============================================================
function Invoke-PreparePackedAssets {
    Write-Step 'Preparing embedded assets...'

    $sourceDir = Join-Path $Script:BackendDir 'manifest'
    $targetDir = Join-Path $Script:BackendDir 'internal\packed\manifest'

    # Rebuild the packed manifest workspace from scratch
    if (Test-Path $targetDir) {
        Remove-Item -Recurse -Force $targetDir
    }
    Ensure-Dir "$targetDir\config"
    Ensure-Dir "$targetDir\sql"
    Ensure-Dir "$targetDir\i18n"

    # Copy only distributable manifest assets (exclude local config.yaml)
    Copy-Item (Join-Path $sourceDir 'config\config.template.yaml') "$targetDir\config\" -ErrorAction Stop
    Copy-Item (Join-Path $sourceDir 'config\metadata.yaml') "$targetDir\config\" -ErrorAction Stop
    Copy-Item (Join-Path $sourceDir 'sql\*') "$targetDir\sql\" -Recurse -ErrorAction Stop
    Copy-Item (Join-Path $sourceDir 'i18n\*') "$targetDir\i18n\" -Recurse -ErrorAction Stop

    # Placeholder .gitkeep
    '' | Out-File -FilePath (Join-Path $targetDir '.gitkeep') -Encoding utf8

    Write-Success 'Embedded assets prepared'
}

# ============================================================
# Command: dev
# ============================================================
function Invoke-Dev {
    Write-Banner
    Write-Step 'Starting LinaPro development environment...'

    Invoke-Stop

    # Preflight: verify frontend dependencies are installed
    $frontendNM = Join-Path $Script:FrontendDir 'apps\web-antd\node_modules'
    if (!(Test-Path $frontendNM)) {
        Write-Step 'Frontend dependencies not found. Running pnpm install...'
        Push-Location $Script:FrontendDir
        try {
            pnpm install
            if ($LASTEXITCODE -ne 0) { throw 'pnpm install failed' }
        }
        finally { Pop-Location }
        Write-Success 'Frontend dependencies installed'
    }

    Ensure-Dir $Script:TempDir
    Ensure-Dir $Script:PidDir
    Ensure-Dir (Join-Path $Script:TempDir 'bin')

    # Ensure backend config.yaml exists (copy from template on first run)
    $configTarget = Join-Path $Script:BackendDir 'manifest\config\config.yaml'
    if (!(Test-Path $configTarget)) {
        $configTemplate = Join-Path $Script:BackendDir 'manifest\config\config.template.yaml'
        if (Test-Path $configTemplate) {
            Copy-Item $configTemplate $configTarget
            Write-Info "Created config.yaml from template"
        }
    }

    '' | Out-File -FilePath $Script:BackendLog -Encoding utf8
    '' | Out-File -FilePath $Script:FrontendLog -Encoding utf8

    Invoke-Wasm

    Invoke-PreparePackedAssets

    $backendBinary = Join-Path $Script:TempDir 'bin\lina.exe'
    Write-Step 'Building backend...'
    Push-Location $Script:BackendDir
    try {
        $output = & go build -o $backendBinary . 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Host $output
            throw 'Backend build failed'
        }
    }
    finally { Pop-Location }
    Write-Success 'Backend built'

    $procStyle = if ($Background) { @{ WindowStyle = 'Hidden' } } else { @{ NoNewWindow = $true } }

    Write-Step 'Starting backend...'
    $backendProc = Start-Process -FilePath $backendBinary `
        -WorkingDirectory $Script:BackendDir `
        @procStyle `
        -RedirectStandardOutput $Script:BackendLog `
        -RedirectStandardError (Join-Path $Script:TempDir 'lina-core.err.log') `
        -PassThru
    $backendProc.Id | Out-File -FilePath $Script:BackendPidFile -NoNewline
    Write-Info "Backend PID: $($backendProc.Id)"

    Write-Step 'Starting frontend Vite dev server...'
    $frontendWorkDir = Join-Path $Script:FrontendDir 'apps\web-antd'
    $viteArgs = "pnpm exec vite --mode development --host 127.0.0.1 --port $($Script:FrontendPort) --strictPort"
    $frontendProc = Start-Process -FilePath 'cmd' `
        -ArgumentList @('/c', $viteArgs) `
        -WorkingDirectory $frontendWorkDir `
        @procStyle `
        -RedirectStandardOutput $Script:FrontendLog `
        -RedirectStandardError (Join-Path $Script:TempDir 'lina-vben.err.log') `
        -PassThru
    $frontendProc.Id | Out-File -FilePath $Script:FrontendPidFile -NoNewline
    Write-Info "Frontend PID: $($frontendProc.Id)"

    Wait-HttpReady 'Backend' $Script:BackendPidFile "http://127.0.0.1:$($Script:BackendPort)/api/v1/health/get" 60 $Script:BackendLog
    Wait-HttpReady 'Frontend' $Script:FrontendPidFile "http://127.0.0.1:$($Script:FrontendPort)/" 60 $Script:FrontendLog

    Invoke-Status

    if ($Background) {
        Write-Info 'Services running in background. Close this terminal safely. Use stop to shut down.'
    }
    else {
        Write-Host ''
        Write-Host 'Services running. Press Ctrl+C to stop all services.' -ForegroundColor Cyan
        try {
            while ($true) { Start-Sleep -Seconds 1 }
        }
        finally {
            Write-Host ''
            Invoke-Stop
        }
    }
}

# ============================================================
# Command: stop
# ============================================================
function Invoke-Stop {
    Write-Step 'Stopping services...'
    Stop-Service 'Backend' $Script:BackendPidFile $Script:BackendPort
    Stop-Service 'Frontend' $Script:FrontendPidFile $Script:FrontendPort
}

# ============================================================
# Command: status
# ============================================================
function Invoke-Status {
    Write-Host ''
    Write-Host '================================================' -ForegroundColor Cyan
    Write-Host '       LinaPro Framework Status' -ForegroundColor Cyan
    Write-Host '================================================' -ForegroundColor Cyan

    if (Test-PortInUse -Port $Script:BackendPort) {
        Write-Host "  Backend:   RUNNING  http://localhost:$($Script:BackendPort)" -ForegroundColor Green
    }
    else {
        Write-Host "  Backend:   STOPPED  (port $($Script:BackendPort))" -ForegroundColor Red
    }

    if (Test-PortInUse -Port $Script:FrontendPort) {
        Write-Host "  Frontend:  RUNNING  http://localhost:$($Script:FrontendPort)" -ForegroundColor Green
    }
    else {
        Write-Host "  Frontend:  STOPPED  (port $($Script:FrontendPort))" -ForegroundColor Red
    }

    Write-Host '------------------------------------------------' -ForegroundColor Cyan
    Write-Host "  Backend log:   $Script:BackendLog" -ForegroundColor DarkGray
    Write-Host "  Frontend log:  $Script:FrontendLog" -ForegroundColor DarkGray
    Write-Host '================================================' -ForegroundColor Cyan
    Write-Host ''
}

# ============================================================
# Command: build
# ============================================================
function Invoke-Build {
    Write-Banner
    Write-Step 'Starting production build...'

    Write-Step 'Reading build environment...'
    $env = Get-BuildEnv
    $buildOutputDir = if ($env.ContainsKey('BUILD_OUTPUT_DIR')) { $env['BUILD_OUTPUT_DIR'] } else { $Script:OutputDirDefault }
    $buildPlatformsStr = if ($env.ContainsKey('BUILD_PLATFORMS')) { $env['BUILD_PLATFORMS'] } else { 'linux/amd64' }
    $buildPlatforms = $buildPlatformsStr -split '\s+'
    $buildMultiPlatform = if ($env.ContainsKey('BUILD_MULTI_PLATFORM')) { $env['BUILD_MULTI_PLATFORM'] -eq '1' } else { $false }
    $buildCgoEnabled = if ($env.ContainsKey('BUILD_CGO_ENABLED')) { $env['BUILD_CGO_ENABLED'] } else { '0' }
    $buildBinaryName = if ($env.ContainsKey('BUILD_BINARY_NAME')) { $env['BUILD_BINARY_NAME'] } else { 'lina' }
    $buildBinaryPath = if ($env.ContainsKey('BUILD_BINARY_PATH')) { $env['BUILD_BINARY_PATH'] } else { Join-Path $buildOutputDir $buildBinaryName }

    Write-Info "Output dir: $buildOutputDir"
    Write-Info "Platforms: $buildPlatformsStr"
    Write-Info "Binary: $buildBinaryName"

    if (Test-Path $buildOutputDir) { Remove-Item -Recurse -Force $buildOutputDir }
    Ensure-Dir $buildOutputDir

    Write-Step 'Building frontend...'
    Push-Location $Script:FrontendDir
    try {
        if ($VerboseOutput) {
            pnpm run build
        }
        else {
            pnpm run build 2>&1 | Out-Null
        }
        if ($LASTEXITCODE -ne 0) { throw 'Frontend build failed' }
    }
    finally { Pop-Location }
    Write-Success 'Frontend built'

    Write-Step 'Copying frontend assets to embed directory...'
    $frontendDist = Join-Path $Script:FrontendDir 'apps\web-antd\dist'
    if (Test-Path $Script:EmbedDir) { Remove-Item -Recurse -Force $Script:EmbedDir }
    Ensure-Dir $Script:EmbedDir
    Copy-Item -Path "$frontendDist\*" -Destination $Script:EmbedDir -Recurse
    Write-Success 'Frontend embedded assets generated'

    Invoke-PreparePackedAssets

    Invoke-Wasm -OutDir $buildOutputDir
    Write-Success 'Dynamic plugin artifacts generated'

    foreach ($targetPlatform in $buildPlatforms) {
        $os, $arch = $targetPlatform -split '/'
        if (!$os) { $os = 'linux' }
        if (!$arch) { $arch = 'amd64' }

        if ($buildMultiPlatform) {
            $targetBinaryPath = Join-Path $buildOutputDir "$($os)_$arch\$buildBinaryName"
        }
        else {
            $targetBinaryPath = $buildBinaryPath
        }

        Ensure-Dir (Split-Path $targetBinaryPath -Parent)
        Write-Step "Building backend for $os/$arch..."

        Push-Location $Script:BackendDir
        try {
            $env:CGO_ENABLED = $buildCgoEnabled
            $env:GOOS = $os
            $env:GOARCH = $arch

            if ($VerboseOutput) {
                go build -o $targetBinaryPath .
            }
            else {
                go build -o $targetBinaryPath . 2>&1 | Out-Null
            }
            if ($LASTEXITCODE -ne 0) { throw "Backend build failed for $os/$arch" }
        }
        finally { Pop-Location }
        Write-Success "Build complete: $targetBinaryPath"
    }

    Write-Success 'Full build completed successfully'
}

# ============================================================
# Command: wasm
# ============================================================
function Invoke-Wasm {
    param([string]$OutDir)
    Write-Step 'Building dynamic WASM plugins...'

    $builderDir = Join-Path $Script:RootDir 'hack\tools\build-wasm'
    $outputDir = if ($OutDir) { $OutDir } elseif ($OutputDir) { $OutputDir } else { $Script:OutputDirDefault }

    if (![System.IO.Path]::IsPathRooted($outputDir)) {
        $outputDir = Join-Path (Get-Location) $outputDir
    }
    Ensure-Dir $outputDir

    $pluginsDir = Join-Path $Script:RootDir 'apps\lina-plugins'

    if ($Plugin) {
        $pluginDir = Join-Path $pluginsDir $Plugin
        if (!(Test-Path $pluginDir)) {
            throw "Plugin not found: $Plugin"
        }
        $manifestFile = Join-Path $pluginDir 'plugin.yaml'
        if (!(Test-Path $manifestFile)) {
            throw "plugin.yaml not found in $Plugin"
        }
        Write-Info "Building: $Plugin"
        Push-Location $builderDir
        try {
            go run . --plugin-dir $pluginDir --output-dir $outputDir
        }
        finally { Pop-Location }
        Write-Success "Plugin built: $Plugin"
    }
    else {
        $dynamicPlugins = @()
        $pluginDirs = Get-ChildItem -Path $pluginsDir -Directory -ErrorAction SilentlyContinue
        foreach ($dir in $pluginDirs) {
            $manifest = Join-Path $dir.FullName 'plugin.yaml'
            if (Test-Path $manifest) {
                $content = Get-Content $manifest -Raw
                if ($content -match 'type:\s*dynamic') {
                    $dynamicPlugins += $dir.Name
                }
            }
        }

        if ($dynamicPlugins.Count -eq 0) {
            Write-Info 'No buildable dynamic WASM plugins found'
            return
        }

        foreach ($plugin in $dynamicPlugins) {
            Write-Info "Building: $plugin"
            $pluginDir = Join-Path $pluginsDir $plugin
            Push-Location $builderDir
            try {
                go run . --plugin-dir $pluginDir --output-dir $outputDir
            }
            finally { Pop-Location }
        }
        Write-Success "Built $($dynamicPlugins.Count) dynamic plugin(s)"
    }
}

# ============================================================
# Command: init
# ============================================================
function Invoke-Init {
    if ($Confirm -ne 'init') {
        Write-Host '[ERROR] init requires explicit confirmation for safety' -ForegroundColor Red
        Write-Host '  Use: .\dev.ps1 init -Confirm init'
        Write-Host '  To rebuild: .\dev.ps1 init -Confirm init -Rebuild true'
        exit 1
    }

    Write-Step 'Initializing database...'
    Push-Location $Script:BackendDir
    try {
        $rebuildArg = if ($Rebuild) { "--rebuild=$Rebuild" } else { '' }
        $cmdResult = & go run main.go init --confirm=$Confirm --sql-source=local $rebuildArg 2>&1
        $exitCode = $LASTEXITCODE

        if ($exitCode -ne 0) {
            Write-Host $cmdResult
            if ($cmdResult -match 'dial tcp|connection refused|connect: connection|failed to connect|i/o timeout|no such host') {
                Write-Host ''
                Write-Host 'PostgreSQL is not ready. Start PostgreSQL first.' -ForegroundColor Yellow
                Write-Host 'Local example:' -ForegroundColor DarkGray
                Write-Host '  docker run --name linapro-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=linapro -p 5432:5432 -d postgres:14-alpine' -ForegroundColor DarkGray
            }
            throw 'Database initialization failed'
        }
        Write-Host $cmdResult
    }
    finally { Pop-Location }
    Write-Success 'Database initialization complete'
}

# ============================================================
# Command: mock
# ============================================================
function Invoke-Mock {
    if ($Confirm -ne 'mock') {
        Write-Host '[ERROR] mock requires explicit confirmation for safety' -ForegroundColor Red
        Write-Host '  Use: .\dev.ps1 mock -Confirm mock'
        exit 1
    }

    Write-Step 'Loading mock data...'
    Push-Location $Script:BackendDir
    try {
        go run main.go mock --confirm=$Confirm --sql-source=local
        if ($LASTEXITCODE -ne 0) { throw 'Mock data load failed' }
    }
    finally { Pop-Location }
    Write-Success 'Mock data load complete'
}

# ============================================================
# Command: test (E2E)
# ============================================================
function Invoke-Test {
    Write-Step 'Running E2E test suite...'
    Push-Location (Join-Path $Script:RootDir 'hack\tests')
    try {
        pnpm test
        if ($LASTEXITCODE -ne 0) { throw 'E2E tests failed' }
    }
    finally { Pop-Location }
}

# ============================================================
# Command: test-go
# ============================================================
function Invoke-TestGo {
    Write-Step 'Running Go unit tests with race detector...'
    $moduleDirs = & go list -m -f '{{.Dir}}' 2>&1
    if ($LASTEXITCODE -ne 0) { throw 'Failed to list Go modules' }
    if ([string]::IsNullOrWhiteSpace($moduleDirs)) {
        throw 'No Go workspace modules discovered'
    }

    foreach ($moduleDir in $moduleDirs) {
        $moduleDir = $moduleDir.Trim()
        if ([string]::IsNullOrWhiteSpace($moduleDir)) { continue }
        Write-Step "go test -race -v $moduleDir"
        Push-Location $moduleDir
        try {
            go test -race -v ./...
            if ($LASTEXITCODE -ne 0) { throw "Tests failed in $moduleDir" }
        }
        finally { Pop-Location }
    }
    Write-Success 'All Go tests passed'
}

# ============================================================
# Command: check-i18n
# ============================================================
function Invoke-CheckI18n {
    Write-Step 'Scanning for hard-coded text...'
    go run ./hack/tools/runtime-i18n scan
    if ($LASTEXITCODE -ne 0) { throw 'i18n scan failed' }
}

# ============================================================
# Command: check-i18n-messages
# ============================================================
function Invoke-CheckI18nMessages {
    Write-Step 'Validating i18n message key coverage...'
    go run ./hack/tools/runtime-i18n messages
    if ($LASTEXITCODE -ne 0) { throw 'i18n message check failed' }
}

# ============================================================
# Command: image
# ============================================================
function Invoke-Image {
    Write-Banner
    $ibArgs = Get-ImageBuilderArgs

    Write-Step 'Running image preflight checks...'
    go run ./hack/tools/image-builder --preflight $ibArgs
    if ($LASTEXITCODE -ne 0) { throw 'Image preflight failed' }

    Write-Step 'Building artifacts for Docker image...'
    Invoke-Build

    Write-Step 'Building Docker image...'
    go run ./hack/tools/image-builder $ibArgs
    if ($LASTEXITCODE -ne 0) { throw 'Image build failed' }
    Write-Success 'Docker image built successfully'
}

# ============================================================
# Command: image-build
# ============================================================
function Invoke-ImageBuild {
    Write-Step 'Staging image build artifacts...'
    $ibArgs = Get-ImageBuilderArgs

    Invoke-Build

    go run ./hack/tools/image-builder --build-only $ibArgs
    if ($LASTEXITCODE -ne 0) { throw 'Image artifact staging failed' }
    Write-Success 'Image artifacts staged successfully'
}

# ============================================================
# GoFrame code generation commands
# ============================================================
function Invoke-GfGen {
    param([string]$GenCommand)

    Push-Location $Script:BackendDir
    try {
        $gfCheck = Get-Command gf -ErrorAction SilentlyContinue
        if ($null -eq $gfCheck) {
            Write-Host '[ERROR] GoFrame CLI (gf) is not installed.' -ForegroundColor Red
            Write-Host '  Download from: https://github.com/gogf/gf/releases/latest'
            Write-Host '  Select the Windows .exe and add to PATH.'
            throw 'gf CLI not found'
        }

        Write-Step "Running: gf gen $GenCommand"
        Invoke-Expression "gf gen $GenCommand"
        if ($LASTEXITCODE -ne 0) { throw "gf gen $GenCommand failed" }
    }
    finally { Pop-Location }
    Write-Success "gf gen $GenCommand completed"
}

# ============================================================
# Command: check
# ============================================================
function Invoke-Check {
    Write-Banner
    Write-Host 'Checking development environment...' -ForegroundColor Cyan
    Write-Host ''

    $allOk = $true
    $issues = @()

    # ---- PowerShell version ----
    $psVer = $PSVersionTable.PSVersion
    if ($psVer.Major -ge 5) {
        Write-Success "PowerShell: v$psVer"
    }
    else {
        Write-Host "  [MISS] PowerShell $psVer is too old (need 5.1+)" -ForegroundColor Red
        $allOk = $false
        $issues += 'Upgrade PowerShell to 5.1+. Windows 10+ ships with it.'
    }

    # ---- Execution policy ----
    $execPolicy = Get-ExecutionPolicy -Scope CurrentUser -ErrorAction SilentlyContinue
    if ($execPolicy -eq 'Restricted' -or $execPolicy -eq 'AllSigned') {
        Write-Host "  [WARN] Execution policy is $execPolicy — scripts will be blocked" -ForegroundColor Yellow
        Write-Host '          Run: Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser' -ForegroundColor DarkGray
        $issues += 'Run: Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser'
    }
    else {
        Write-Success "ExecPolicy: $execPolicy"
    }

    # ---- Go ----
    $goPath = Get-Command go -ErrorAction SilentlyContinue
    if ($goPath) {
        $goVer = & go version 2>&1
        Write-Success "Go:         $goVer"
    }
    else {
        Write-Host '  [MISS] Go is not in PATH' -ForegroundColor Red
        $allOk = $false
        $issues += 'Install Go 1.25+ from https://go.dev/dl/'
    }

    # ---- Node.js ----
    $nodePath = Get-Command node -ErrorAction SilentlyContinue
    if ($nodePath) {
        $nodeVer = & node --version 2>&1
        Write-Success "Node.js:    $nodeVer"
    }
    else {
        Write-Host '  [MISS] Node.js is not in PATH' -ForegroundColor Red
        $allOk = $false
        $issues += 'Install Node.js LTS from https://nodejs.org/'
    }

    # ---- pnpm ----
    $pnpmPath = Get-Command pnpm -ErrorAction SilentlyContinue
    if ($pnpmPath) {
        $pnpmVer = & pnpm --version 2>&1
        Write-Success "pnpm:       v$pnpmVer"
    }
    else {
        Write-Host '  [MISS] pnpm is not in PATH' -ForegroundColor Red
        $allOk = $false
        $issues += 'Run: corepack enable (if using Node.js built-in) or: npm install -g pnpm'
    }

    # ---- Git Bash (optional, for remaining shell scripts) ----
    $bashPath = Get-Command bash -ErrorAction SilentlyContinue
    if ($bashPath) {
        Write-Success 'Git Bash:   available'
    }
    else {
        Write-Host '  [WARN] Git Bash not found (optional, for legacy shell scripts)' -ForegroundColor Yellow
    }

    # ---- Docker ----
    $dockerCmd = Get-Command docker -ErrorAction SilentlyContinue
    $dockerRunning = $false
    if ($dockerCmd) {
        $dockerVer = & docker --version 2>&1
        Write-Success "Docker:     $dockerVer"
        # Check if Docker daemon is actually running
        try {
            $dockerInfo = & docker info 2>&1
            if ($LASTEXITCODE -eq 0) {
                $dockerRunning = $true
            }
            else {
                Write-Host '  [WARN] Docker is installed but daemon is not running' -ForegroundColor Yellow
                Write-Host '          Start Docker Desktop first' -ForegroundColor DarkGray
            }
        }
        catch {
            Write-Host '  [WARN] Docker is installed but daemon is not running' -ForegroundColor Yellow
            Write-Host '          Start Docker Desktop first' -ForegroundColor DarkGray
        }
    }
    else {
        Write-Host '  [WARN] Docker not found (optional, needed for local PostgreSQL)' -ForegroundColor Yellow
        Write-Host '          Install Docker Desktop from https://www.docker.com/products/docker-desktop/' -ForegroundColor DarkGray
    }

    # ---- PostgreSQL database connectivity ----
    $pgPort = 5432
    $pgReachable = $false
    try {
        $tcp = New-Object System.Net.Sockets.TcpClient
        if ($tcp.ConnectAsync('127.0.0.1', $pgPort).Wait(2000)) {
            $pgReachable = $true
            $tcp.Close()
        }
    }
    catch { }

    if ($pgReachable) {
        Write-Success "PostgreSQL: reachable on 127.0.0.1:$pgPort"
    }
    else {
        # Check if container exists but is stopped
        $containerStatus = $null
        if ($dockerRunning) {
            try {
                $containerStatus = & docker ps -a --filter name=linapro-postgres --format '{{.Status}}' 2>&1
            }
            catch { }
        }

        if ($containerStatus) {
            if ($containerStatus -match '^Up') {
                Write-Info "PostgreSQL: container is running but port $pgPort not reachable"
                Write-Host '          Check docker port mapping: -p 5432:5432' -ForegroundColor DarkGray
            }
            else {
                Write-Host "  [WARN] PostgreSQL container exists but is stopped: $containerStatus" -ForegroundColor Yellow
                Write-Host '          Start it: docker start linapro-postgres' -ForegroundColor DarkGray
            }
        }
        else {
            Write-Host "  [WARN] PostgreSQL not reachable on 127.0.0.1:$pgPort" -ForegroundColor Yellow
            if ($dockerCmd) {
                if ($dockerRunning) {
                    Write-Host '          Start a local instance:' -ForegroundColor DarkGray
                    Write-Host '          docker run --name linapro-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=linapro -p 5432:5432 -d postgres:14-alpine' -ForegroundColor DarkGray
                }
                else {
                    Write-Host '          Start Docker Desktop, then:' -ForegroundColor DarkGray
                    Write-Host '          docker run --name linapro-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=linapro -p 5432:5432 -d postgres:14-alpine' -ForegroundColor DarkGray
                }
            }
            else {
                Write-Host '          Install Docker first, or use a remote PostgreSQL instance' -ForegroundColor DarkGray
                Write-Host '          Configure connection in manifest/config/config.yaml -> database.default.link' -ForegroundColor DarkGray
            }
        }
    }

    # ---- Frontend node_modules ----
    $frontendNM = Join-Path $Script:FrontendDir 'apps\web-antd\node_modules'
    if (Test-Path $frontendNM) {
        $nmCount = (Get-ChildItem $frontendNM -Directory).Count
        Write-Success "Frontend:   node_modules present ($nmCount packages)"
    }
    else {
        Write-Host '  [WARN] Frontend node_modules missing. Run: .\dev.ps1 dev (auto-installs)' -ForegroundColor Yellow
    }

    # ---- Go workspace modules ----
    Write-Info 'Go workspace modules:'
    $modDirs = & go list -m -f '{{.Dir}}' 2>$null
    if ($LASTEXITCODE -eq 0 -and $modDirs) {
        foreach ($m in $modDirs) {
            if ($m.Trim()) {
                Write-Info "  $($m.Trim())"
            }
        }
    }

    # ---- Ports ----
    Write-Info ''
    Write-Info 'Port status:'
    if (Test-PortInUse -Port $Script:BackendPort) {
        Write-Host "  [WARN] Backend port $($Script:BackendPort) is in use -- will conflict with dev" -ForegroundColor Yellow
    }
    else {
        Write-Success "Backend port $($Script:BackendPort) is free"
    }
    if (Test-PortInUse -Port $Script:FrontendPort) {
        Write-Host "  [WARN] Frontend port $($Script:FrontendPort) is in use -- will conflict with dev" -ForegroundColor Yellow
    }
    else {
        Write-Success "Frontend port $($Script:FrontendPort) is free"
    }

    # ---- GoFrame CLI (optional) ----
    $gfPath = Get-Command gf -ErrorAction SilentlyContinue
    if ($gfPath) {
        Write-Success 'gf CLI:     available (for code generation)'
    }
    else {
        Write-Info 'gf CLI:     not installed (optional, for code generation only)'
    }

    # ---- Go environment ----
    Write-Info ''
    $goEnv = & go env GOPATH GOROOT GOVERSION GOOS GOARCH 2>&1
    foreach ($line in $goEnv) {
        Write-Info "  $line"
    }

    # ---- Summary ----
    Write-Host ''
    if ($allOk) {
        Write-Host '================================================' -ForegroundColor Green
        Write-Host '  Environment check passed. Ready to develop!' -ForegroundColor Green
        Write-Host '  Run: .\dev.ps1 dev' -ForegroundColor Green
        Write-Host '================================================' -ForegroundColor Green
    }
    else {
        Write-Host '================================================' -ForegroundColor Red
        Write-Host '  Issues found:' -ForegroundColor Red
        foreach ($issue in $issues) {
            Write-Host "    - $issue" -ForegroundColor Yellow
        }
        Write-Host '================================================' -ForegroundColor Red
    }
    Write-Host ''
}

# ============================================================
# Command: clean
# ============================================================
function Invoke-Clean {
    Write-Step 'Cleaning temp files...'
    if (Test-Path $Script:TempDir) {
        Remove-Item -Recurse -Force $Script:TempDir -ErrorAction SilentlyContinue
    }
    Write-Success 'Temp files cleaned'
}

# ============================================================
# Command: help
# ============================================================
function Invoke-Help {
    Write-Banner
    Write-Host 'Usage: .\dev.ps1 <command> [options]' -ForegroundColor DarkGray
    Write-Host ''
    Write-Host 'Development Commands:' -ForegroundColor Cyan
    Write-Host '  dev                 Start full-stack dev server'
    Write-Host '                      Use -Background to detach from terminal'
    Write-Host '  stop                Stop all running dev services'
    Write-Host '  status              Show backend/frontend runtime status'
    Write-Host ''
    Write-Host 'Build Commands:' -ForegroundColor Cyan
    Write-Host '  build               Full production build (frontend + embed + plugins + backend)'
    Write-Host '                      Options: -Platforms linux/amd64 -VerboseOutput'
    Write-Host '  wasm                Build dynamic WASM plugins (all, or use -Plugin <plugin-id>)'
    Write-Host ''
    Write-Host 'Database Commands:' -ForegroundColor Cyan
    Write-Host '  init                Initialize database with DDL and seed data'
    Write-Host '                      Required: -Confirm init'
    Write-Host '                      Optional: -Rebuild true'
    Write-Host '  mock                Load mock demo data (requires init first)'
    Write-Host '                      Required: -Confirm mock'
    Write-Host ''
    Write-Host 'Test Commands:' -ForegroundColor Cyan
    Write-Host '  test                Run Playwright E2E test suite'
    Write-Host '  test-go             Run Go unit tests with race detector'
    Write-Host ''
    Write-Host 'Quality Commands:' -ForegroundColor Cyan
    Write-Host '  check-i18n          Scan for hard-coded text in runtime-visible paths'
    Write-Host '  check-i18n-messages Validate runtime i18n message key coverage'
    Write-Host ''
    Write-Host 'Docker Image Commands:' -ForegroundColor Cyan
    Write-Host '  image               Build production Docker image'
    Write-Host '                      Options: -ImageTag v0.6.0 -Registry ghcr.io/linaproai -Push 1'
    Write-Host '  image-build         Stage image build artifacts (without docker build)'
    Write-Host ''
    Write-Host 'Code Generation Commands (require gf CLI):' -ForegroundColor Cyan
    Write-Host '  dao                 Generate DAO/DO/Entity Go files'
    Write-Host '  ctrl                Generate controllers and SDK from API definitions'
    Write-Host '  service             Generate Service layer Go files'
    Write-Host '  enums               Generate enum Go files'
    Write-Host '  pb                  Generate protobuf Go files'
    Write-Host '  pbentity            Generate protobuf entities from database tables'
    Write-Host ''
    Write-Host 'Other Commands:' -ForegroundColor Cyan
    Write-Host '  check               Verify development environment is ready'
    Write-Host '  clean               Clean temp directory'
    Write-Host '  help                Show this help message'
    Write-Host ''
    Write-Host 'First-time setup:' -ForegroundColor DarkGray
    Write-Host '  .\dev.ps1 check                          Verify dev environment' -ForegroundColor DarkGray
    Write-Host '  docker run ... postgres:14-alpine         Start PostgreSQL' -ForegroundColor DarkGray
    Write-Host '  .\dev.ps1 init -Confirm init             Initialize database' -ForegroundColor DarkGray
    Write-Host '  .\dev.ps1 mock -Confirm mock             Load demo data' -ForegroundColor DarkGray
    Write-Host ''
    Write-Host 'Daily development:' -ForegroundColor DarkGray
    Write-Host '  .\dev.ps1 dev                            Start dev server' -ForegroundColor DarkGray
    Write-Host '  .\dev.ps1 status                         Check status' -ForegroundColor DarkGray
    Write-Host '  .\dev.ps1 stop                           Stop services' -ForegroundColor DarkGray
    Write-Host ''
    Write-Host 'Other:' -ForegroundColor DarkGray
    Write-Host '  .\dev.ps1 build -VerboseOutput           Production build' -ForegroundColor DarkGray
    Write-Host '  .\dev.ps1 wasm -Plugin plugin-demo-dynamic Build single plugin' -ForegroundColor DarkGray
    Write-Host '  .\dev.ps1 test-go                        All Go unit tests' -ForegroundColor DarkGray
    Write-Host ''
}

# ============================================================
# Main Dispatch
# ============================================================

if ($Help -or (-not $Command)) {
    Invoke-Help
    exit 0
}

try {
    switch ($Command) {
        'dev'                   { Invoke-Dev }
        'stop'                  { Invoke-Stop }
        'status'                { Invoke-Status }
        'build'                 { Invoke-Build }
        'wasm'                  { Invoke-Wasm }
        'init'                  { Invoke-Init }
        'mock'                  { Invoke-Mock }
        'test'                  { Invoke-Test }
        'test-go'               { Invoke-TestGo }
        'check-i18n'            { Invoke-CheckI18n }
        'check-i18n-messages'   { Invoke-CheckI18nMessages }
        'image'                 { Invoke-Image }
        'image-build'           { Invoke-ImageBuild }
        'dao'                   { Invoke-GfGen 'dao' }
        'ctrl'                  { Invoke-GfGen 'ctrl' }
        'service'               { Invoke-GfGen 'service' }
        'enums'                 { Invoke-GfGen 'enums' }
        'pb'                    { Invoke-GfGen 'pb' }
        'pbentity'              { Invoke-GfGen 'pbentity' }
        'clean'                 { Invoke-Clean }
        'check'                 { Invoke-Check }
        'help'                  { Invoke-Help }
    }
}
catch {
    Write-Host ''
    Write-Host "  [ERROR] $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ''
    exit 1
}
