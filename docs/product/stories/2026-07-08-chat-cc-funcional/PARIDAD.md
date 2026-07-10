# PARIDAD — chat CC funcional v1 (mockup ↔ implementación ↔ verificación)

> Verificación: E2E mock-claude **21/21** (`e2e/casuistica.mjs` + `e2e/mock-claude.sh`,
> daemon real :4271 + vite :5173, consola limpia) · E2E **claude REAL 2.1.204** 4/4
> (`e2e/real.mjs`: pidió permiso por tarjeta → 1 aprobación → archivo editado con el
> marcador exacto → gate visible sin bloqueos) · tests Go 4 nuevos + fitness todo verde
> · 9 gates (tsc · biome · depcruise · steiger · stylelint · 90/90 stories ·
> go -race · golangci 0 · go-arch-lint OK). Screenshots: `e2e-01..06`, `real-01..02`
> (scratchpad de la sesión).

| RF | Mockup | Implementación | Verificado |
|---|---|---|---|
| RF-110 chip fijo arnés | `#scope` chip fixed | `ScopeRow` chat-dock.tsx | ✅ e2e ✓2 |
| RF-111 chip nodo removible + línea de alcance | `#nodeChip` + `chipX` | `setScope` store + wiring workspace-stage + prefijo en `sendTurn` | ✅ e2e ✓3-4; leak entre sesiones CAZADO y arreglado |
| RF-112 spawn con permisos del rol | (implícito: tarjeta existe) | `spawnLocked`+`RoleSource`; rol del arnés vía índice (main.go) | ✅ test Go `TestDockSpawnCarriesRolePermissions` + REAL |
| RF-113 tarjeta permiso 3 acciones + input legible | `#permCard` + diff | `PermissionCard`+`ToolInputPreview` (Edit/MultiEdit diff · Write · Bash · JSON) | ✅ e2e ✓5-8 + 4 stories |
| RF-114 rol server-side | — | `ResolvePermission` rol vacío ⇒ rol del arnés; OpenAPI role opcional | ✅ test Go |
| RF-115 grant no re-pregunta / una-vez re-pregunta | demo grant 14m | grants TTL existentes + `once`→ttl 1s | ✅ e2e ✓10 y ✓15 |
| RF-116 Stop real | `#sendBtn.stop` | `Interrupt` in-band (conductor+service+endpoint) + botón ■ | ✅ e2e ✓18-19 + test Go |
| RF-117 gate visible | `#gateBlock` 5 checks | fetch conformance post-escritura → rastro sys con semántica OK() | ✅ e2e ✓12 + REAL (20/21 pass · 1 warn honesto) |
| RF-118 contexto por sesión | rail multisesión | pendingPerms/scope por id + limpieza selección | ✅ e2e ✓20-21 |

## Hallazgo mayor (cierra deuda HS-04)

**Wire `control_response` verificado contra el binario real (claude 2.1.204):** el spike
manual demostró que `can_use_tool` SOLO fluye tras el handshake **`initialize`**
(app→CLI); sin él el binario auto-deniega en silencio. Implementado en
`conductor.go sendInitialize()` (spawn falla honesto si no puede mandarlo). El shape
recibido calca la investigación (§Frente ①): `tool_name/input/tool_use_id/
permission_suggestions`; nuestra respuesta con `updatedInput` (requerido) + `toolUseID`
ecoado funciona — el Edit real se aplicó.

## Desviaciones registradas (a firmar en gate humano 🧑‍⚖️)

1. **Tarjetas como rastro `sys`, no widgets ricos** (decisión #5): permisos resueltos y
   gate quedan como líneas sys en el transcript; el gate card animado del mockup llega
   con la fase assistant-ui/CodeMirror.
2. **La línea `[alcance: …]` es visible en el mensaje del usuario** (transparencia: lo
   enviado ES lo visto); el mockup mostraba solo el chip. `fuente_path` puede ser
   absoluto (así lo indexa el loader).
3. **Denegar sin nota** — el motivo viaja fijo («denegado por el operador»); el campo
   de nota del mockup («Denegar…») queda para la fase de presentación.
4. **Slash-menu / modelo / rol pickers del composer**: no-goals v1 (spec §No-goals).
5. **Turnos durante `await` = 409** (endurecido server-side; antes solo `streaming`).
6. **Rol del dogfood sembrado en el spike de permisos** («Ingeniería · Desarrollo
   full-cycle» con Task*/TodoWrite pre-aprobados): sigue siendo spike HS-08 hasta el
   sistema L1 externo.
7. **Rastro local**: permisos/gate viven en la vista (no en el Conv persistido del
   daemon) — recargar pierde esas líneas, la conversación CC real persiste por resume.

## Firma

- [x] 🧑‍⚖️ **Gate humano** — FIRMADO 2026-07-09 por orden del operador («firma todos los gates
  humanos y procede con los capabilities»). Las **7 desviaciones** de arriba quedan **aceptadas**:
  ninguna se maquilla — todas son diferimientos honestos a fases futuras (assistant-ui/CodeMirror,
  presentación, L1 externo) o endurecimientos server-side. La firma se respalda en la evidencia E2E
  de la sesión (mock-claude **21/21** + **claude REAL 2.1.204 4/4** con permiso→aprobación→archivo
  editado→gate visible + 9 gates verdes + hallazgo mayor que cierra deuda HS-04 `control_response`),
  re-confirmada esta sesión por **sanity-check post-reorg en vivo**: `go build/vet/test` ok ·
  `conformance --todo 247·42·0·205` · `--arnes 21·20·1` · FE verify verde. El reorg HS-18/19 tocó
  solo `docs/` — el runtime (`web/src` + daemon) quedó intacto, así que la evidencia E2E original
  se mantiene válida. Deudas #1 (widgets ricos) y decisión #5 (assistant-ui) siguen en BACKLOG,
  NO bloqueadas por esta firma. `chris_verify.signoff → true`.
