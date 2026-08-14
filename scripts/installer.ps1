# installer.ps1 - espejo Windows del target `_installer-build` del Makefile
# (paquete 2026-08-13-compilacion-windows). Produce los instaladores de escritorio
# (.msi via WiX + .exe via NSIS) y los deja versionados en instaladores/vX.Y.Z/.
#
# El Makefile no corre en Windows (SHELL := /usr/bin/env bash + find/sha256sum/
# install/pkill). Este script replica su SEMANTICA exacta con utilidades nativas:
#   1. guard: una generacion publicada NUNCA se pisa (falla si vX.Y.Z/ tiene archivos)
#   2. limpia el bundle dir (artefactos viejos jamas viajan a una version nueva)
#   3. scripts/bundle.py  (SPA + daemon sellado RF-231 + sidecar .exe + tauri build)
#   4. copia *.msi / *.exe a instaladores/vX.Y.Z/
#   5. checksums.txt con Get-FileHash SHA256 (reemplaza sha256sum)
#
# NO hace dev-sync: en Windows no existe el override ~/.local/bin/arnesia que el
# shell prefiere en Linux (decision DA-9: en Windows siempre el sidecar empaquetado).
#
# Sin bump: este script empaqueta la version ACTUAL de los manifiestos (equivale a
# `make installer-actual`). El bump sigue siendo `bash scripts/bump.sh` en Linux
# (bump.py portable = deuda registrada en el BACKLOG).
#
# Nota de encoding: ASCII puro a proposito - PowerShell 5.1 lee UTF-8-sin-BOM como
# ANSI y los acentos saldrian corruptos en la consola.
#
# Uso:  powershell -File scripts\installer.ps1

$ErrorActionPreference = 'Stop'

$Root = Split-Path -Parent $PSScriptRoot
$CargoToml = Join-Path $Root 'web\src-tauri\Cargo.toml'
$InstallDir = Join-Path $Root 'instaladores'

# --- version: misma fuente de verdad y misma regex que bundle.py::version_de_cargo ---
$m = Select-String -Path $CargoToml -Pattern '^version = "(.*)"' | Select-Object -First 1
if (-not $m) { throw "no pude leer la version de $CargoToml" }
$Version = $m.Matches[0].Groups[1].Value
$Destino = Join-Path $InstallDir "v$Version"

Write-Host "== build instalador v$Version =="

# --- 1. guard: nunca pisar una generacion publicada (igual que el Makefile) ---
if ((Test-Path $Destino) -and (Get-ChildItem $Destino -Force | Measure-Object).Count -gt 0) {
    throw "instaladores/v$Version/ ya tiene archivos - una generacion publicada NUNCA se pisa (bumpea primero)"
}

# --- bundle dir: se resuelve por cargo metadata (robusto ante CARGO_TARGET_DIR) ---
$SrcTauri = Join-Path $Root 'web\src-tauri'
$TargetDir = $null
try {
    Push-Location $SrcTauri
    $meta = (& cargo metadata --format-version 1 --no-deps 2>$null) | ConvertFrom-Json
    if ($meta.target_directory) { $TargetDir = $meta.target_directory }
} catch {
    Write-Host "aviso: cargo metadata no resolvio el target dir; uso la ruta estandar"
} finally {
    Pop-Location
}
if (-not $TargetDir) { $TargetDir = Join-Path $SrcTauri 'target' }
$BundleDir = Join-Path $TargetDir 'release\bundle'

# --- 2. bundle dir limpio ---
if (Test-Path $BundleDir) { Remove-Item $BundleDir -Recurse -Force }

# --- 3. el bundle real (mismo script que usa el self-update para sellar) ---
# python3 en Windows suele ser el stub falso de WindowsApps (existe en PATH y no ejecuta):
# se sondea EJECUTANDO, igual que el shim PYTHON del Makefile.
$py = 'python'
foreach ($cand in 'python3', 'python') {
    try {
        & $cand -c "pass" 2>$null
        if ($LASTEXITCODE -eq 0) { $py = $cand; break }
    } catch { }
}
& $py (Join-Path $Root 'scripts\bundle.py')
if ($LASTEXITCODE -ne 0) { throw "bundle.py fallo con exit $LASTEXITCODE" }

# --- 4. recoger los instaladores producidos ---
New-Item -ItemType Directory -Force -Path $Destino | Out-Null
$artefactos = Get-ChildItem $BundleDir -Recurse -File -Include *.msi, *.exe -ErrorAction SilentlyContinue
if (-not $artefactos) {
    throw "bundle no genero ningun instalador reconocido en $BundleDir - revisar tauri.conf.json bundle.targets"
}
$artefactos | ForEach-Object { Copy-Item $_.FullName $Destino -Force }

# --- 5. checksums (sha256sum no existe en Windows) ---
Push-Location $Destino
try {
    Get-ChildItem -File | Where-Object Name -ne 'checksums.txt' | ForEach-Object {
        "{0}  {1}" -f (Get-FileHash $_.FullName -Algorithm SHA256).Hash.ToLower(), $_.Name
    } | Set-Content 'checksums.txt' -Encoding ascii
} finally {
    Pop-Location
}

Write-Host "OK - instalador v$Version en $Destino"
Get-ChildItem $Destino | Select-Object Name, @{n='MB';e={[math]::Round($_.Length/1MB,2)}} | Format-Table -AutoSize
Write-Host "recorda commitear el CHANGELOG + los manifiestos de version (instaladores/ esta gitignoreado)"
