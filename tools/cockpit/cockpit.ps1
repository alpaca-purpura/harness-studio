# cockpit.ps1 - build / run / stop / status del cockpit vendored, en Windows.
#
# El Makefile del repo no corre aca (SHELL := /usr/bin/env bash), y build.sh/build-ui.sh
# del cockpit son bash. Este script replica su semantica con utilidades nativas.
#
# El build tiene un paso no obvio: Next.js con output:'export' NO admite rutas API, pero
# el arbol ui/app/api existe (es el cockpit Next original, del que el server Go es un
# port). Se aparta durante el build y se restaura SIEMPRE - incluso si el build falla,
# porque perderlas dejaria el arbol upstream mutilado.
#
# Nota de encoding: ASCII puro a proposito - PowerShell 5.1 lee UTF-8-sin-BOM como ANSI
# y los acentos saldrian corruptos en la consola. (Mismo criterio que scripts/installer.ps1.)
#
# Uso:  powershell -File tools\cockpit\cockpit.ps1 build|run|stop|status

param(
    [Parameter(Position = 0)]
    [ValidateSet('build', 'run', 'stop', 'status')]
    [string]$Accion = 'status',

    [int]$Port = 4300
)

$ErrorActionPreference = 'Stop'

$CockpitDir = $PSScriptRoot
$GoDir      = Join-Path $CockpitDir 'go'
$UiDir      = Join-Path $CockpitDir 'ui'
$Root       = Split-Path -Parent (Split-Path -Parent $CockpitDir)
$LocalDir   = Join-Path $CockpitDir '.cockpit-local'
$PidFile    = Join-Path $LocalDir 'cockpit.pid'
$LogFile    = Join-Path $LocalDir 'cockpit.log'
$Exe        = Join-Path $GoDir 'cockpit.exe'

function Get-CockpitPid {
    if (-not (Test-Path $PidFile)) { return $null }
    $raw = (Get-Content $PidFile -Raw).Trim()
    if (-not $raw) { return $null }
    $procId = 0
    if (-not [int]::TryParse($raw, [ref]$procId)) { return $null }
    try { $null = Get-Process -Id $procId -ErrorAction Stop } catch { return $null }
    return $procId
}

function Invoke-Build {
    Write-Host "== build cockpit =="

    # --- 1. deps de la UI (--ignore-workspace: el cockpit es un arbol npm INDEPENDIENTE
    #        del pnpm-workspace del repo host; sin el flag pnpm intenta enlazarlo) ---
    Push-Location $UiDir
    try {
        if (-not (Test-Path (Join-Path $UiDir 'node_modules'))) {
            Write-Host "-- pnpm install"
            & pnpm install --ignore-workspace
            if ($LASTEXITCODE -ne 0) { throw "pnpm install fallo con exit $LASTEXITCODE" }
        }

        # --- 2. export estatico, con las rutas API apartadas ---
        $ApiDir   = Join-Path $UiDir 'app\api'
        $ApiStash = Join-Path $UiDir '.api-stash'
        # Restos de una corrida anterior interrumpida: el stash es la copia buena.
        if ((Test-Path $ApiStash) -and (Test-Path $ApiDir)) { Remove-Item $ApiDir -Recurse -Force }
        if (Test-Path $ApiDir) { Move-Item $ApiDir $ApiStash }
        try {
            Remove-Item (Join-Path $UiDir '.next') -Recurse -Force -ErrorAction SilentlyContinue
            Remove-Item (Join-Path $UiDir 'out')   -Recurse -Force -ErrorAction SilentlyContinue
            Write-Host "-- next build (COCKPIT_STATIC=1)"
            $env:COCKPIT_STATIC = '1'
            & pnpm exec next build
            if ($LASTEXITCODE -ne 0) { throw "next build fallo con exit $LASTEXITCODE" }
        } finally {
            # SIEMPRE: si el build revienta, las rutas API vuelven igual.
            if (Test-Path $ApiStash) {
                if (Test-Path $ApiDir) { Remove-Item $ApiDir -Recurse -Force }
                Move-Item $ApiStash $ApiDir
            }
            Remove-Item Env:\COCKPIT_STATIC -ErrorAction SilentlyContinue
        }
    } finally {
        Pop-Location
    }

    # --- 3. la UI compilada al lugar donde el go:embed la busca ---
    $EmbedDir = Join-Path $GoDir 'ui'
    if (Test-Path $EmbedDir) { Remove-Item $EmbedDir -Recurse -Force }
    Copy-Item (Join-Path $UiDir 'out') $EmbedDir -Recurse

    # --- 4. el binario, sellado con la version del repo host (misma fuente que
    #        bundle.py e installer.ps1: web\src-tauri\Cargo.toml) ---
    $CargoToml = Join-Path $Root 'web\src-tauri\Cargo.toml'
    $Version = 'dev'
    $m = Select-String -Path $CargoToml -Pattern '^version = "(.*)"' | Select-Object -First 1
    if ($m) { $Version = $m.Matches[0].Groups[1].Value }

    Push-Location $GoDir
    try {
        & go build -ldflags="-s -w -X main.version=$Version" -o cockpit.exe .
        if ($LASTEXITCODE -ne 0) { throw "go build fallo con exit $LASTEXITCODE" }
    } finally {
        Pop-Location
    }

    $mb = [math]::Round((Get-Item $Exe).Length / 1MB, 2)
    Write-Host "OK - cockpit.exe v$Version ($mb MB)"
}

function Invoke-Run {
    if (-not (Test-Path $Exe)) { throw "no existe $Exe - corre primero: cockpit.ps1 build" }
    $existente = Get-CockpitPid
    if ($existente) { throw "cockpit ya corriendo (pid $existente) - cockpit.ps1 stop primero" }

    New-Item -ItemType Directory -Force -Path $LocalDir | Out-Null

    # -allow-platform-writes: harness-studio tiene su docs/product en la RAIZ, o sea que
    # TODO su contenido es el pseudo-sistema `platform`. Sin el flag el board seria de
    # solo lectura y el drag daria 400 (ver workspace.go::writableSistemas).
    $argumentos = @('start', '-port', "$Port", '-workspace', $Root, '-allow-platform-writes')

    $proc = Start-Process -FilePath $Exe -ArgumentList $argumentos `
        -RedirectStandardOutput $LogFile -RedirectStandardError "$LogFile.err" `
        -WindowStyle Hidden -PassThru
    Set-Content $PidFile $proc.Id -Encoding ascii

    Start-Sleep -Milliseconds 900
    if ($proc.HasExited) {
        Write-Host "x el cockpit murio al arrancar - log:"
        if (Test-Path "$LogFile.err") { Get-Content "$LogFile.err" -Tail 20 }
        Remove-Item $PidFile -ErrorAction SilentlyContinue
        exit 1
    }

    Write-Host "OK - cockpit corriendo (pid $($proc.Id)) - http://localhost:$Port"
    Write-Host "     workspace: $Root"
    Write-Host "     log:       $LogFile"
}

function Invoke-Stop {
    $procId = Get-CockpitPid
    if (-not $procId) {
        Write-Host "cockpit no esta corriendo"
        Remove-Item $PidFile -ErrorAction SilentlyContinue
        return
    }
    Stop-Process -Id $procId -Force
    Remove-Item $PidFile -ErrorAction SilentlyContinue
    Write-Host "OK - cockpit detenido (pid $procId)"
}

function Invoke-Status {
    $procId = Get-CockpitPid
    if ($procId) {
        Write-Host "cockpit CORRIENDO (pid $procId) - http://localhost:$Port"
    } else {
        Write-Host "cockpit detenido"
    }
    if (Test-Path $Exe) {
        $mb = [math]::Round((Get-Item $Exe).Length / 1MB, 2)
        Write-Host "binario: $Exe ($mb MB, $((Get-Item $Exe).LastWriteTime))"
    } else {
        Write-Host "binario: AUSENTE - corre cockpit.ps1 build"
    }
}

switch ($Accion) {
    'build'  { Invoke-Build }
    'run'    { Invoke-Run }
    'stop'   { Invoke-Stop }
    'status' { Invoke-Status }
}
