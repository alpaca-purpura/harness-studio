# Decisiones — definición canónica de arnés

> Registro same-turn (disciplina §10). Firmas vía AskUserQuestion 2026-07-30.

- **DEF-D1 — FIRMADA 🧑‍⚖️ 2026-07-30 · feature/bugfix/spike = TIPOS de paquete de
  trabajo con pipeline propio DENTRO del mismo arnés-dev, NO arneses separados.**
  Aplica D20 del modelo de terreno (multi-pipeline por tipo, firmada 2026-07-10) al
  caso dev concreto: mismo rol → mismo arnés; cada tipo declara
  `{jerarquia, estados, criterio_cierre}` — el spike cierra en aprendizaje
  documentado (jamás producción), el bugfix salta discovery, la feature corre el
  spine completo. Regla mnemotécnica ratificada: **¿cambia QUIÉN opera? → otro
  arnés; ¿cambia solo la RUTA? → otro pipeline del mismo arnés.**
- **DEF-D2 — FIRMADA 🧑‍⚖️ 2026-08-01 · Definición canónica v3 + vocabulario.**
  Tras 2 rounds de debate (v1 rechazada-para-debate → v2 → aporte del operador en
  `chris-input.md` round 2 → v3):
  > **Arnés** = el paquete de know-how de **UN puesto** sobre **UN tipo de
  > terreno**: agrupa las **ACTIVIDADES** del puesto dentro de su tramo del
  > proceso (su macroproceso), cada actividad con su **procedimiento** (pipeline)
  > propio sobre una **base común**. 1 arnés = 1 plugin; **se instala POR
  > CARPETA** — cada instalación opera un terreno concreto. El proceso end-to-end
  > es entidad aparte (DEF-D3); el corte proceso→arneses y la agrupación
  > actividades→arnés se rigen por el criterio de corte. 2 roles → 2 arneses.

  **Vocabulario firmado:** la palabra «arnés» vive a nivel PAQUETE/plugin (como
  todo lo construido: sello, `(home,id)`, Portafolio, Mapa, `--arnes`); el nivel
  interno = **actividad** con su **procedimiento**. **Criterio de corte firmado:**
  ¿otro operador? → separá SIEMPRE (seguridad) · ¿otro terreno? → probablemente ·
  ¿solo otra actividad, mismo operador+terreno? → JAMÁS (pipeline del mismo
  arnés). Jerarquía completa y mapeo → `debate-definicion.md` §4b.
- **DEF-D3 — FIRMADA 🧑‍⚖️ 2026-08-01 · La cadena = `proceso` entidad de primer
  orden en el home.** `proceso/<id>.yaml` declarado en el marketplace de la
  familia (SSoT por ley anti-drift): lista de fases, cada una con arnés-owner +
  `gate_salida` VERIFICABLE (el gate de salida de una fase = contrato de entrada
  de la siguiente). META se afila: `proceso` string → referencia `proceso: <id>` +
  `fases: [...]`. Mapeo fases→arnés N:1-friendly (empresa chica: un arnés posee
  todas las fases; crecer = reasignar, no re-modelar). Galaxia renderiza la cadena
  desde el home; telemetría joinea por `proceso.id`. Visibilidad cross-rol vive
  acá (Mapa/Galaxia read-only), no en los plugins. Diseño → `debate-definicion.md`
  §5. **Quedan a la spec** (no bloquean la firma): ¿copia offline del tramo en
  cada `arnes.yaml`? · ¿versionado del proceso independiente del semver de cada
  arnés?
- **DEF-D5 — DIRECTIVA del operador 2026-08-01 · Mapa por actividades + spec ya.**
  «Que no se vea saturado: ver el todo, seleccionar una actividad y ver su proceso,
  pero todo conectado — hay aspectos que van más allá de un paso y afectan a nivel
  proyecto (arquitectura, etc.)». Bajado a spec (`spec.md`: leyes MA-L1..L7,
  niveles N0/N1, escenarios MA-E1..E12) e **integrado al MVP como carril D**.
  El orden §10 mockup→spec se invirtió por directiva; el mockup (MA-T6) queda como
  gate 🧑‍⚖️ previo a construir FE. **La spec espera firma 🧑‍⚖️.**
- **DEF-D6 — DIRECTIVA del operador 2026-08-01 · dogfood = `developer-vitalia`.**
  Extraer el primer arnés del esquema v3 de `/home/chalreme/Proyectos/vitalia-app`
  (material CRUDO: 42 skills · 10 agents · 45 rules · 4 hooks · 10 commands · 1
  workflow, sin sello en raíz). **Eliminar la ficción de marketplace** (donde el
  Portafolio la muestre como-si-de-marketplace → material crudo `sin-home`).
  Corte por puesto developer (primera aplicación real del criterio de corte:
  pm/po/sales/brand quedan FUERA, material de futuros arneses hermanos), 4
  actividades (historia · bugfix · spike · revisar-capability), sellar y publicar
  al marketplace propio como plugin **`developer-vitalia`** (write-side B2).
  Detalle → `spec.md` §5.
- **DEF-D8 — FIRMA 🧑‍⚖️ 2026-08-01 · spec v2 + gate del mockup MA-T6, juntas.**
  Transcripta del /goal del operador («ok firmo, ejecuta y desarrolla todo con
  tus propios mecanismos de verificación y validación…»). Directivas adjuntas a
  la firma: arnés de prueba = `~/Proyectos/vitalia-app` (sin marketplace ni
  plugin aún) · la forja de `developer-vitalia` se hace **CONVERSACIONALMENTE
  DESDE LA APP** («para que ya comencemos a usar el app correctamente») ·
  verificación/validación propias del constructor · ante dudas: preguntar,
  jamás asumir. Desbloquea MA-T1a..T7 completo.
- **AUD — registro de auditoría de la spec contra el sistema real (2026-08-01,
  pedido del operador: «auditala, revisala y completala»; spec → v2, FIRMADA
  vía DEF-D8).** Hallazgos aplicados como correcciones:
  - **AUD-1 · la cadena paso→caja NO existía como dato** (ningún `PasoSpine`
    referencia cajas — `forja/parser.go:47-53`): la spec ahora la crea como campo
    opcional `caja:` por paso + escenario E13 (paso sin caja VISIBLE).
  - **AUD-2 · el seam no está cableado**: el daemon jamás lee `arnes.yaml` de un
    proyecto (gap Fase 2 declarado, `parser.go:6-7`); el payload del Mapa es
    passthrough del índice. MA-T1 se partió en T1a (dato) + T1b (seam + derivar
    AL INDEXAR, patrón edges).
  - **AUD-3 · schema L0 raíz cerrada** (`additionalProperties:false`): catálogo
    va en `arnes.actividades[]` vía enmienda ADITIVA; faceta por caja ya legal.
    Drift destapado de paso: `degradado` en wire y no en schema → misma enmienda.
  - **AUD-4 · «spine» tomado en el wire** (= máquina de estados singular): la
    secuencia por actividad se llama `pasos`/procedimiento; en FE jamás `act`.
  - **AUD-5 · vocabulario real de la semilla**: `cierre` (no `criterio_cierre`),
    id = clave del mapa, `label` no existe; 13 stories (no 14); inspector 4 tabs.
  - **AUD-6 · MA-L7 sin dato de puertos**: la satisface la franja de artefactos
    vigente (`externo`/`salida del proceso`); `proceso/<id>.yaml` no existe aún.
  - **AUD-7 · saneo vitalia — estado observado ≠ asumido**: lo indexado ya está
    `sin-home` (y es `luana-vitalia/vitalia`, no vitalia-app); lo real a sanear =
    clave legacy `vitalia` duplicada (residuo re-key) + vitalia-app jamás
    escaneado. MA-T2 redefinido.
  - **AUD-8 · corte §5.2 completado** contra inventario real (42·10·10·45·4·1;
    clerk-*=12 exacto): DENTRO/FUERA explícitos + 6 ambiguos que decide el
    operador en el gate de MA-T3 (architect · auditor · manychat-expert ·
    handoff · harnesses-improvement · harness-issue). `pase-produccion` FUERA
    por el ejemplo 1 del operador.
  - **AUD-9 · N0/N1 anclados a lo vigente**: salud del chip = worst-of hallazgos
    FE existentes (sin dato nuevo); foco N1 reusa `related`/`.dim`
    (`map-canvas.tsx:101-171`); secuencia = sólido `--primary` numerado (los dash
    ya significan escribe/lee/opcional); breadcrumb junto al id read-only
    (picker eliminado, TS-D21). Baseline `.html` declarado STALE → MA-T6
    re-deriva a la superficie vigente (4 tabs · node-meta · PRENTER · charset).
- **DEF-D4 — registro (hallazgo, no decisión): gentle-ai analizado 2026-07-30.**
  NO es un plugin CC ni contiene arneses según nuestra definición: es un
  configurador de ecosistema transversal (14 agentes, memoria Engram, SDD opcional,
  receipts RDD, skill registry), outcome-first, **sin modelo de rol ni de proceso
  de negocio**. No compite con el arnés — compite con el harness crudo. Ideas a
  evaluar: instalación modular con scope por workspace · «receipt» único que
  validan todas las compuertas (nosotros: PARIDAD + conformance, más fuerte pero
  menos ergonómico). Detalle → `debate-definicion.md` §6.
