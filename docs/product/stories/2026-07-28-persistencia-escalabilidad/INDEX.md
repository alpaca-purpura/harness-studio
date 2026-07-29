# Persistencia: correctness + escalabilidad (paquete de trabajo)

> `tipo: mejora backend, sin UI` · **Vig: construido y verificado, falta firma 🧑‍⚖️ de PARIDAD.**
> Detonante: auditoría propia de la capa de persistencia a pedido del operador + `/goal`
> "implementa todas las fases 1 a 1... probando vos mismo... contra la app instalada".

## El agujero (auditado, no supuesto)

- `portafolio.json`/`marketplaces.json` sin backup/cuarentena (a diferencia de `sessions.json`, que
  sí tiene) — la doctrina llama a `portafolio.json` "identidad más irrecuperable" y es el store con
  MENOS protección de los cuatro.
- Sin lock cross-proceso: `runPortafolio`/`cmdSesiones` (CLI standalone) abren su propia instancia
  del store, independiente de la que vive en `runServe` (daemon) — dos cachés en memoria escribiendo
  el mismo archivo, last-writer-wins confirmado por lectura de código.
- `portafolio.json`↔`index.db` puede divergir en silencio (ley anti-drift S1-D1 es correcta en
  intención, pero no expone la brecha).

Ver [`decisiones.md`](./decisiones.md) para el detalle + por qué el plan de Fase 1 se revisó en
vivo (flock, no proxy HTTP con un token nuevo).

## Fases

- [x] **Fase 0** — backup/cuarentena real + version-check (Modo E) en `portafolio.Store`/
      `marketplace.Store`. Tests: `TestStoreArchivoTotalmenteIlegible` (reescrito para probar
      cuarentena real, no pérdida), `TestStoreEsquemaFuturoEsSoloLectura` (los 2 stores).
      Verificado en vivo: portafolio.json roto → daemon arranca, cuarentena en disco, sirve vacío.
- [x] **Fase 1** — `internal/adapters/filelock` (flock/LockFileEx cross-proceso, sin dependencia
      nueva) + reload-guarded en `portafolio.Store.Upsert/Desvincular`,
      `marketplace.Store.Upsert/Olvidar`, `ArnesRegistry.Register` + guard de ficha-viva en
      `arnesia sesiones recalibrar-llaves` (`--force` para saltarlo). Tests: 2 de proceso-real
      (`TestGuardSerializesAcrossProcesses`, re-exec de `os.Args[0]`) + 2 regresión a nivel Store
      (`TestUpsertNoPisaEscrituraDeOtraInstancia`, `TestArnesRegistryRegisterNoPisaEscrituraDeOtraInstancia`,
      confirmadas FALLANDO al deshabilitar el fix a propósito, y pasando con él). Verificado en
      vivo contra el binario compilado: 12 `agregar` concurrentes (mitad HTTP-daemon, mitad
      CLI-standalone) → 12/12 sobrevivieron.
- [x] **Fase 2** — `avisarCoberturaPortafolioIndice` en el boot de `runServe`: cuenta entradas del
      Portafolio sin observar en el Mapa, `slog.Warn` con la cuenta + claves (silencioso si
      cobertura completa). Verificado contra una COPIA real del `~/.arnesia` del operador: 3 de 4
      entradas reales resultaron sin observar — hallazgo real, no sintético.
- [x] Mecanismo de prueba promovido a skill: `.claude/skills/verificando-binario-instalado/`
      (skill de apoyo, `contract.caja: false`) — build sandboxed + HOME-copia + HTTP+CLI cruzados
      + kill-por-ruta-exacta + concurrencia real por re-exec + cierre contra el binario instalado.
- [x] `go test ./...` completo (incluye `docs/architecture/fitness` — boundaries + R1/R2 capability
      coverage) + verificación manual repetida contra sandbox Y contra `~/.local/bin/arnesia` real
      tras `make dev-sync` (identidad `0.4.0.2607282209`): boot real disparó la migración
      pendiente `sessions.json→sesiones.json` (con su respaldo `.bak`, como siempre) + el re-key
      CV-D18 + el aviso nuevo de Fase 2 — todo correcto, nada inesperado.
- [x] `CHANGELOG.md` — 4 entradas en `Agregado`, una por pieza, mismo turno.
- [x] Capability yaml actualizado: `indice-persistencia/persistencia.yaml` (filelock + Register) +
      `portafolio/observar-en-mapa.yaml` (aviso de cobertura) — `cap_doctor.py` + R1/R2 en verde.
- [ ] 🧑‍⚖️ **Gate de PARIDAD** — pendiente, firma del operador.

Fuera de alcance (Fase 3, explícita): migrar los JSON a SQL por volumen — no aplica a un daemon
local mono-usuario con cardinalidad de decenas de arneses. Ver `decisiones.md`.

## Hallazgo lateral (no de este paquete, FYI)

`~/.config/arnesia/daemon.json` tiene una ficha huérfana fechada 2026-07-27 14:48 (pid 37956,
ya no vivo) — quedó de una sesión real anterior que no cerró ordenado. No la tocamos (no es
nuestra). Consecuencia práctica: `daemonPareceVivo()` (Fase 1, guard de `recalibrar-llaves`) va a
dar un falso positivo hasta que se borre a mano o el daemon real vuelva a arrancar y la
sobreescriba — el operador necesitaría `--force` si corre ese comando hoy contra la máquina.
