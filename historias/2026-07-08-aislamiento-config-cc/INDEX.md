# Aislamiento de superficie de config al spawnear CC (paquete de trabajo)

> Ficha [HS-17](../../LEDGER.md) (diagnóstico) · 2026-07-08
> Origen: `/context` de una sesión interactiva mostró 62.2k tok en MCP ajenos (Canva/
> Gmail/Google Calendar/Google Drive/claude_design) — disparador que destapó un eje real
> del producto: cada arnés que ArnesIA spawnea hoy hereda `~/.claude/settings.json`,
> CLAUDE.md ascendente y MCP de cuenta del OPERADOR que corre el daemon, encima de su
> propio kit ①②. Objetivo: que cada arnés cargue SOLO lo que ArnesIA inyecta — «solo los
> arneses propios» (VISION §Constitución) llevado al eje de config-surface, no solo al de
> contenido.
> Disciplina METODOLOGIA §10. **Adaptación explícita:** paquete 100% backend/infra, CERO
> superficie UI — no aplica `mockup-*.html` ni `design.md` (precedente de trabajo backend-
> puro como `.md` suelto: `arquitectura-fase3.md`/`arquitectura-inyeccion-knowhow.md`; acá
> se usa carpeta completa igual, por la regla de continuidad multi-sesión que pidió el
> operador — «anotalo... para continuar esto en múltiples conversaciones»). Flujo real:
> investigación → decisiones → spec → implementación por fases → PARIDAD técnica (gates
> verdes + verificación en vivo, sin click-through visual porque no hay UI) → cierre HS-17.

## Rumbo firmado

Nada firmado aún por el operador. Lo que SÍ está resuelto por evidencia dura (no requiere
firma, es hecho técnico): **`--bare` descartado** — rompe auth de suscripción (fuerza
`ANTHROPIC_API_KEY`/`apiKeyHelper`, ya documentado en knowledge v1.1 desde HS-04, re-
confirmado con cita nueva 2026-07-08). El resto (qué flags SÍ se adoptan, cómo se cementan
as-code) está en `decisiones.md` como PROPUESTA, pendiente de firma antes de tocar código.

## Flujo y gates (adaptado — sin mockup/design, backend puro)

1. **Investigación** — 3 subagentes `claude-code-guide` contra `code.claude.com/docs`
   (revisado 2026-07-08). Hecho, ver `spec.md` §Evidencia.
2. **Decisiones** (`decisiones.md`) — cada elección técnica, PROPUESTA→FIRMADA. Al
   instante de conversarse, nunca al final.
3. **Spec** (`spec.md`) — RF-160..RF-16x, cada uno trazado a evidencia (URL+cita) en vez
   de `mockup:línea` (no hay mockup). → 🧑‍⚖️ firma del paquete antes de codear.
4. **Implementación** por fases (0 baseline → 1 MCP → 2 settings-sources → 3 experimento
   `--safe-mode` → 4 cementado as-code → 5 gate de calidad) — cada fase resultado
   verificado EN VIVO contra el dogfood real, nunca simulado (disciplina del repo).
5. **Cierre** — HS-17 pasa de diagnóstico a ejecutada (mismo patrón HS-14): ficha de
   cierre en LEDGER.md con verificación real.

## Estado

- [x] Diagnóstico inicial (conversación previa, ahora ficha HS-17 en LEDGER.md)
- [x] Investigación 3-frentes verificada (fuentes oficiales, citas en `spec.md`)
- [x] Plan detallado de 6 fases propuesto (esta conversación)
- [x] Paquete de trabajo creado (este INDEX + decisiones.md + spec.md)
- [x] `decisiones.md` firmado por el operador (2026-07-08, vía prompt de arranque)
- [x] `spec.md` firmado por el operador → implementación AUTORIZADA
- [x] Fase 0 — baseline real (sonda contra dogfood, SIN cambios de código) —
      `baseline.md`: confirma la fuga — MCP de cuenta (chrome-devtools/Canva/Gmail/
      Calendar/claude_design, "todavía conectando") + ~50 skills `golang-*` +
      plugins `caveman`/stitch-design ajenos, todos visibles en el spawn del dogfood
- [x] Fase 1 — aislamiento MCP (`--mcp-config`+`--strict-mcp-config` siempre) —
      `Injection.MCPConfigFile` + `Provisioner` materializa `~/.arnesia/mcp.json` +
      `SpawnArgs` agrega los flags; `TestSpawnArgsMCPAislado` +
      `TestProvisionMaterializesAndIsIdempotent` (extendido) verdes; sonda real
      `post-mcp.md`: 6→0 MCP de cuenta, skills propias intactas;
      `conformance --arnes` dogfood sin regresión (20 pass/1 warn-fail, igual que antes)
- [x] Fase 2 — `--setting-sources project,local` (excluir `user`) — `SpawnArgs` lo fija
      incondicional (sin campo `Injection`, nadie puede reintroducir `user`);
      `TestSpawnArgsSettingSourcesExcludeUser` verde; sonda real `post-settings.md`:
      ~50 skills `golang-*` + plugins `caveman`/stitch AJENOS desaparecen, kit propio
      (`arnesia-kit:*`) intacto, CLAUDE.md del arnés sobrevive (cita textual exacta,
      confirma E5). Hallazgo residual DOCUMENTADO (esperado por D3, no bloqueante): el
      walk ascendente de CLAUDE.md sigue activo — en este dev-env el dogfood anidado
      bajo el repo hereda el CLAUDE.md raíz del producto; en producción normal un arnés
      no vive anidado así. `conformance --arnes` sin regresión (20 pass/1 warn-fail)
- [ ] Fase 3 — experimento `--safe-mode` (decidir adopción según resultado, no a priori)
- [ ] Fase 4 — cementado as-code (extensión de `superficie-local-confinada` + fitness test)
- [ ] Fase 5 — gate de calidad (round-trip dogfood + `conformance --arnes` sin regresión +
      medición de contexto antes/después)
- [ ] Fase 6 — cierre: ficha de cierre HS-17 en LEDGER.md

## Retomar aquí

- **Último hecho (2026-07-08):** investigación de 3 subagentes (`--safe-mode` vs
  `--bare` · `--setting-sources` + CLAUDE.md discovery · `--strict-mcp-config` + MCP en
  plugins) verificada contra `code.claude.com/docs` (revisado 2026-07-08) → plan de 6
  fases presentado al operador → operador pide fichar TODO antes de tocar código, para
  poder continuar en otra conversación. Este paquete es esa anotación.
- **Próximo paso concreto:** el operador revisa `decisiones.md` y `spec.md` de este
  paquete y firma (o pide cambios). Con la firma, arranca Fase 0 (sonda baseline contra
  el dogfood real, sin tocar código todavía) y luego Fase 1 (MCP — el cambio de menor
  riesgo, cero fricción con auth, listo para codear primero).
- **Firmas pendientes:** `decisiones.md` completo · `spec.md` completo (una sola firma
  de paquete, como en `boton-actualizar`/`inspector-drawer`).
- **Contexto caliente para quien retome:** el hallazgo más frágil del plan es que
  `--setting-sources` NO tiene confirmación oficial explícita de si corta
  `enabledPlugins` a nivel `user` (inferencia fuerte, no cita literal) — la Fase 0/2
  DEBEN verificarlo con una sonda real antes de cementar nada as-code. Igual con si
  `--safe-mode` mata `--plugin-dir` explícito o no — sin cita dura, solo experimento
  real lo resuelve (Fase 3). No asumir ninguno de los dos sin la sonda.
