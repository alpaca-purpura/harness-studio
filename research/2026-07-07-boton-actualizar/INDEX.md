# Botón «Actualizar» — self-update sin sudo (paquete de trabajo)

> Ficha HS-11 (deuda del instalador) · 2026-07-07
> Origen: el operador instala cada build con `sudo dpkg -i` (el .deb escribe en
> /usr/bin, dueño root — ESE es el sudo). Pide un botón en la propia app.
> Disciplina METODOLOGIA §10: mockup → decisiones → spec/design → 🧑‍⚖️ → código.

## Rumbo firmado (2026-07-07, respuestas del operador)

Self-update del **binario instalado** · mecanismo **`~/.local/bin`** (espacio de
usuario, cero sudo; el .deb queda para instalación inicial de terceros) · origen
**build local del repo** (el ciclo real de dogfood). Detalle en `decisiones.md`.

## Flujo y gates

1. **Mockup** (`mockup-actualizar.html`) — vista Ajustes con la sección «Versión y
   actualización»; tokens DTCG reales; datos REALES (huella del repo, ruta de
   instalación actual /usr/bin root); estados honestos. Iterar → 🧑‍⚖️ firma.
2. **Decisiones** (`decisiones.md`) — cada cambio conversado, al instante.
3. **Specs** (`spec.md` + `design.md`) → 🧑‍⚖️ firma del paquete.
4. **Implementación** — story=test + tests Go; gates verdes; el endpoint queda BAJO el
   confinamiento S1 (Host+Origin+token) y jamás acepta paths del request.
5. **Paridad** (`PARIDAD.md`) → gate final.

## Estado

- [x] rumbo firmado (alcance · mecanismo · origen)
- [x] mockup **v2 FIRMADO** (operador «firmo el mockup», 2026-07-07) — **artifact
      (publicar SIEMPRE a esta URL):** https://claude.ai/code/artifact/458c147a-b78b-461c-a54b-db2d741ac30b
- [x] spec.md (RF-100..107) + design.md — 🧑‍⚖️ **PAQUETE FIRMADO** (operador
      «Ok, firmo», 2026-07-07) → implementación AUTORIZADA
- [ ] implementación + stories + tests
- [ ] PARIDAD → gate final

## Retomar aquí

- **Último hecho (2026-07-07, v2):** el operador preguntó «¿dónde estará el botón?» →
  **decisión #5 PROPUESTA**: vista global **Ajustes** (⚙ al pie del rail; hoy ComingSoon
  `global-view.tsx:14`) — la tarjeta la estrena; futuras tarjetas (marketplaces · daemon)
  punteadas. Mockup **v2** publicado (misma URL) con el caso 00 UBICACIÓN: frame del
  shell (rail + pie ⌂/⟳/⚙ con Ajustes activo → vista con la tarjeta). Click-through
  ojo-UI hecho con Playwright headless (el browser MCP seguía tomado): 6 casos renderizan,
  interactividad 00/01 recorre pasos→éxito, consola limpia; screenshot en scratchpad.
- **Base v1:** 5 casuísticas de estados sobre datos reales: binario HOY en
  `/usr/bin/arnesia` (root, .deb) — el diseño migra a `~/.local/bin` (una migración
  inicial, después cero sudo); huella `1c7443f`; el binario NO embebe versión aún
  (pendiente `-ldflags`, se decide en spec).
- **Mockup v2 FIRMADO (2026-07-07)** → specs escritos el mismo turno: `spec.md`
  (RF-100..107: vista Ajustes nace · tarjeta identidad honesta vía GET /api/version ·
  estados no-actualizables · POST /api/self-update con 5 pasos y corte al primer fallo ·
  reinicio re-exec + polling · seguridad cementada: withAuth, cero params, 409) +
  `design.md` (tokens de tarjeta/checklist/pills · máquina de estados · FSD
  features/self-update UI-pura + página transporte · hexagonal ports.SelfUpdater +
  adapter selfupdate + usecase · flag --repo). **Nota clave de honestidad:** la
  checklist «en vivo» del mockup era demo — la app pinta la checklist con los
  veredictos REALES de la respuesta (jamás pasos animados inventados); dicho en spec
  RF-104 y design §A11y.
- **PAQUETE FIRMADO (2026-07-07).** Prompt riguroso de arranque persistido en
  [`PROMPT-implementacion.md`](./PROMPT-implementacion.md) (contexto → diseño técnico
  sellado → implementación Go-primero → validación REAL: el daemon actualizándose a sí
  mismo desde ~/.local/bin). PARIDAD.md esqueleto (8 filas).
- **Próximo paso:** nueva conversación ejecuta ese prompt. Al cerrar: PARIDAD ✅,
  gates verdes, «Retomar aquí» actualizado.
- **Firmas pendientes:** ninguna hasta la paridad (gate final tras implementar).
- **Contexto caliente (candidatos a decisión en la iteración):** ① la ruta del repo
  NO viaja en el request — el daemon la conoce por flag/env/registro explícito
  (superficie-local-confinada); ② el self-update requiere toolchain local (go+pnpm) —
  feature de operador-dev, rotulada; ③ reinicio = write-tmp → rename → re-exec (rename
  sobre binario corriendo es legal en Linux); la UI reconecta por /healthz; ④ migración
  inicial /usr/bin→~/.local/bin: una vez, documentada (PATH: ~/.local/bin precede);
  ⑤ versión visible = huella git embebida por ldflags al build.
