# Chat CC completamente funcional — modificar arneses conversando (paquete de trabajo)

> Ficha HS-09 (fase 5) · Hito 2 · 2026-07-08
> Origen: el operador quiere que el chat del Dock sea COMPLETAMENTE funcional — la
> experiencia de usar Claude Code, pero con nuestros arneses propios especializados en
> crear/modificar arneses según la doctrina. Contexto ligado a la selección: cambiar de
> arnés en el Mapa ⇒ contexto de ese arnés; seleccionar un skill ⇒ alcance ese skill.
> Escenario 1 (este paquete): **modificación de arneses existentes.**

## Flujo y gates (METODOLOGIA §10)

1. **Investigación** (`investigacion.md`) — 4 frentes: GitHub CC-UIs · assistant-ui/
   AG-UI/CodeMirror · repo as-is · doctrina/scoping.
2. **Mockup** (`mockup-chat.html`) — tokens DTCG reales · datos reales del dogfood ·
   se itera con el operador hasta 🧑‍⚖️ firma.
3. **Decisiones** (`decisiones.md`) — PROPUESTA → FIRMADA, con porqué.
4. **Specs** (`spec.md` + `design.md`) — RF trazados a `mockup:línea` → 🧑‍⚖️ firma.
5. **Implementación** → `PARIDAD.md` → 🧑‍⚖️ gate final.

## Estado

- [x] `investigacion.md` COMPLETA (4 frentes, 2026-07-08). Hallazgo mayor: **wire-format
      de `can_use_tool`/`initialize`/`interrupt` verificado en SDK Python + 2 repos**
      (cierra en papel el spike `control_response` de HS-04; falta solo confirmar
      contra el binario real). Validación externa: nuestro conductor = patrón del
      segmento ganador (vibe-kanban/claude-code-chat).
- [x] `mockup-chat.html` v1 — **artifact (publicar SIEMPRE a esta URL):**
      https://claude.ai/code/artifact/76da37eb-fb5c-4c70-9961-39e98cd0fb4e
      Verificado Playwright **24/24 asserts · consola limpia** · screenshots dark/light
      revisados. Demo interactiva: permiso→grant→gate→resumen · deny · chip removible ·
      slash menu con meta-skills del kit · toggle tema.
- [x] decisiones #1–#6 (`decisiones.md`; #5/#6 firmadas de facto por el /goal del
      operador que ordenó implementar)
- [x] `spec.md` (RF-110..118 + Gherkin de casuística)
- [x] **IMPLEMENTADO Y VERIFICADO E2E (2026-07-08, /goal):** Go (Permisos en Dock spawn
      vía rol del arnés · `initialize` handshake · interrupt in-band + endpoint ·
      toolUseID · 409 en await) + FE (tarjeta de permiso con diff · chips de alcance ·
      Stop ■ · gate visible · scoping por sesión). **E2E mock 21/21 · E2E claude REAL
      2.1.204 4/4 (editó el dogfood vía tarjeta) · 9 gates verdes.** Ver `PARIDAD.md`.
- [ ] 🧑‍⚖️ gate final humano: firmar las 7 desviaciones de `PARIDAD.md` + mockup v1
- [ ] fase de presentación (decisión #5): assistant-ui + CodeMirror merge + widgets ricos

## Retomar aquí

- **Último hecho (2026-07-08 noche):** /goal ejecutado — chat COMPLETAMENTE funcional
  para modificar arneses: alcance arnés+nodo, permisos human-in-the-loop con grants
  TTL, deny-del-rol, Stop in-band, gate de conformance visible. **Hallazgo mayor:**
  el wire del control protocol quedó verificado contra el binario real — `can_use_tool`
  SOLO fluye tras el handshake `initialize` (sin él: auto-deny silencioso); cierra la
  deuda-spike de HS-04. Un leak real (chip de alcance cruzando sesiones) lo cazó el
  E2E y quedó arreglado.
- **Próximo paso:** gate final humano sobre `PARIDAD.md` §Desviaciones (7) + firma del
  mockup; luego fase de presentación (assistant-ui/CodeMirror, decisión #5) como
  refactor aditivo sobre el mismo store.
- **Cómo re-correr los E2E:** `e2e/casuistica.mjs` (daemon :4271 con
  `e2e/mock-claude.sh` + vite :5173 con `VITE_ARNESIA_API`) · `e2e/real.mjs` (daemon
  con claude real, copia del dogfood registrada).
- **Firmas pendientes:** mockup v1 · desviaciones PARIDAD (7).

## Contexto que este paquete NO duplica (solo enlaza)

- Multisesión firmada it.14: `mockups/arnesia-shell-A-sessions.html` + UX.md.
- Inyección de doctrina (3 cuerpos): METODOLOGIA §9 ·
  `historias/2026-07-05-arquitectura-inyeccion-knowhow.md`.
- Component-selection chat (assistant-ui + AG-UI + CodeMirror merge): HS-04 +
  gate-0 D3 de `historias/2026-07-06-plan-hito2-doctrina-edicion-showcase/`.
- Nomenclatura clase→ubicación v1.1: `docs/architecture/contracts/nomenclatura-arnes.md`.
- Drawer «Editar conversando» (staged, Fase 3/4):
  `historias/2026-07-07-inspector-drawer/spec.md`.
- Pipeline actual del chat: `internal/adapters/agent/claudecode/conductor.go` ·
  `internal/usecase/session_service.go` · `web/src/widgets/chat-dock/`.
